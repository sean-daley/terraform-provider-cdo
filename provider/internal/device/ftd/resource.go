package ftd

import (
	"context"
	"fmt"

	cdoClient "github.com/CiscoDevnet/terraform-provider-cdo/go-client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *cdoClient.Client
}

type ResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	AccessPolicyName types.String   `tfsdk:"access_policy_name"`
	PerformanceTier  types.String   `tfsdk:"performance_tier"`
	Virtual          types.Bool     `tfsdk:"virtual"`
	Licenses         []types.String `tfsdk:"licenses"`

	AccessPolicyUid  types.String `tfsdk:"access_policy_id"`
	GeneratedCommand types.String `tfsdk:"generated_command"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ftd_device" // TODO: _cloud_ftd_device ?
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Firepower Threat Defense device resource. This allows FTD to be onboarded, updated, and deleted on CDO.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the device. This is a UUID and is automatically generated when the device is created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A human-readable name for the Firewall Threat Defense (FTD). This name must be unique.",
				Required:            true,
			},
			"access_policy_name": schema.StringAttribute{
				MarkdownDescription: "The name of the Cloud FMC access policy that will be used by the FTD",
				Required:            true,
				// TODO: make this optional, and use default access policy when not given
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // TODO: can we change access policy after it is created?
				},
			},
			"performance_tier": schema.StringAttribute{
				MarkdownDescription: "The performance tier of the virtual FTD, if virtual is set to false, this field is ignored.",
				Optional:            true,
				// TODO: validator for performance tier, check valid performance tier is given
				// TODO: ignore changes in this field when virtual is false
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // TODO: can we change performance tier after it is created?
				},
			},
			"virtual": schema.BoolAttribute{
				MarkdownDescription: "Whether this FTD is virtual. If false, performance_tier is ignored",
				Required:            true,
				// TODO: can we change this after created?
				// TODO: default value false
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"licenses": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Comma separated list of licenses of this FTD, it must at least contains the \"BASE\" license.",
				Required:            true,
				// TODO: make this not required, when not given, use BASE license
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(), // TODO: can we modify license after FTD is created?
					// TODO: always sort the licenses so that it is the same order, so that it does not change when order of licenses changes
				},
				// TODO: validate the licenses are valid input.
			},
			"generated_command": schema.StringAttribute{
				MarkdownDescription: "The command to run in the FTD to register itself with Cloud FMC.",
				Computed:            true,
			},
			"access_policy_id": schema.StringAttribute{
				MarkdownDescription: "The id of the access policy used by this FTD.",
				Computed:            true,
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cdoClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cdoClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "read FTD resource")

	// 1. read terraform plan data into the model
	var stateData ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. do read
	if err := Read(ctx, r, &stateData); err != nil {
		resp.Diagnostics.AddError("failed to read FTD resource", err.Error())
		return
	}

	// 3. save data into terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
	tflog.Trace(ctx, "read FTD resource done")
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	tflog.Trace(ctx, "create FTD resource")

	// 1. read terraform plan data into model
	var planData ResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if res.Diagnostics.HasError() {
		return
	}

	// 2. create resource & fill model data
	if err := Create(ctx, r, &planData); err != nil {
		res.Diagnostics.AddError("failed to create FTD resource", err.Error())
		return
	}

	// 3. fill terraform state using model data
	res.Diagnostics.Append(res.State.Set(ctx, &planData)...)
	tflog.Trace(ctx, "create FTD resource done")
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	tflog.Trace(ctx, "update FTD resource")

	// 1. read plan and state data from terraform
	var planData ResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if res.Diagnostics.HasError() {
		return
	}
	var stateData ResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if res.Diagnostics.HasError() {
		return
	}

	// 2. update resource & state data
	if err := Update(ctx, r, &planData, &stateData); err != nil {
		res.Diagnostics.AddError("failed to update FTD resource", err.Error())
		return
	}

	// 3. update terraform state with updated state data
	res.Diagnostics.Append(res.State.Set(ctx, &stateData)...)
	tflog.Trace(ctx, "update FTD resource done")
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	tflog.Trace(ctx, "delete FTD resource")

	// 1. read state data from terraform state
	var stateData ResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if res.Diagnostics.HasError() {
		return
	}

	// 2. delete the resource
	if err := Delete(ctx, r, &stateData); err != nil {
		res.Diagnostics.AddError("failed to delete FTD resource", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}