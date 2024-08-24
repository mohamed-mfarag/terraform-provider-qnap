package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mohamed-mfarag/qnap-client-lib"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &containerResource{}
	_ resource.ResourceWithConfigure = &containerResource{}
)

type ContainerSpecModel struct {
	ID                basetypes.StringValue `tfsdk:"id"`
	Type              basetypes.StringValue `tfsdk:"type"`
	Name              basetypes.StringValue `tfsdk:"name"`
	Image             basetypes.StringValue `tfsdk:"image"`
	IPAddress         basetypes.StringValue `tfsdk:"ipaddress"`
	AutoRemove        basetypes.BoolValue   `tfsdk:"autoremove"`
	Tty               basetypes.BoolValue   `tfsdk:"tty"`
	OpenStdin         basetypes.BoolValue   `tfsdk:"openstdin"`
	Network           basetypes.StringValue `tfsdk:"network"`
	NetworkType       basetypes.StringValue `tfsdk:"networktype"`
	Hostname          basetypes.StringValue `tfsdk:"hostname"`
	LastUpdated       types.String          `tfsdk:"last_updated"`
	Runtime           basetypes.StringValue `tfsdk:"runtime"`
	Privileged        basetypes.BoolValue   `tfsdk:"privileged"`
	RemoveAnonVolumes basetypes.BoolValue   `tfsdk:"removeanonvolumes"`
	Env               basetypes.MapValue    `tfsdk:"env"`
	Labels            basetypes.MapValue    `tfsdk:"labels"`
	Devices           basetypes.ListValue   `tfsdk:"devices"`
	Volumes           basetypes.ListValue   `tfsdk:"volumes"`
	PortBindings      basetypes.ListValue   `tfsdk:"portbindings"`
	Networks          basetypes.ListValue   `tfsdk:"networks"`
	Cpupin            basetypes.ObjectValue `tfsdk:"cpupin"`
	RestartPolicy     basetypes.ObjectValue `tfsdk:"restartpolicy"`
	Cmd               types.List            `tfsdk:"cmd"`
	Entrypoint        basetypes.ListValue   `tfsdk:"entrypoint"`
	DNS               basetypes.ListValue   `tfsdk:"dns"`
	Status            basetypes.StringValue `tfsdk:"status"`
}
type NetworkModel struct {
	ID          basetypes.StringValue `tfsdk:"id"`
	Name        basetypes.StringValue `tfsdk:"name"`
	IPAddress   basetypes.StringValue `tfsdk:"ipaddress"`
	DisplayName basetypes.StringValue `tfsdk:"displayname"`
	MACAddress  basetypes.StringValue `tfsdk:"macaddress"`
	Gateway     basetypes.StringValue `tfsdk:"gateway"`
	NetworkType basetypes.StringValue `tfsdk:"networktype"`
	IsStaticIP  basetypes.BoolValue   `tfsdk:"isstaticip"`
}
type RestartPolicyModel struct {
	Name              basetypes.StringValue `tfsdk:"name" default:"always"`
	MaximumRetryCount basetypes.Int32Value  `tfsdk:"maximumretrycount" default:"0"`
}
type CpupinModel struct {
	CPUIDs basetypes.StringValue `tfsdk:"cpuids" default:""`
	Type   basetypes.StringValue `tfsdk:"type" default:"shared"`
}
type PortBindingsModel struct {
	Host      basetypes.Int32Value  `tfsdk:"host"`
	Container basetypes.Int32Value  `tfsdk:"container"`
	Protocol  basetypes.StringValue `tfsdk:"protocol"`
	HostIP    basetypes.StringValue `tfsdk:"hostip"`
}
type VolumesModel struct {
	Type        basetypes.StringValue `tfsdk:"type"`
	Name        basetypes.StringValue `tfsdk:"name"`
	Container   basetypes.StringValue `tfsdk:"container"`
	Source      basetypes.StringValue `tfsdk:"source"`
	Destination basetypes.StringValue `tfsdk:"destination"`
	Permission  basetypes.StringValue `tfsdk:"permission"`
}
type DevicesModel struct {
	Name       basetypes.StringValue `tfsdk:"name"`
	Permission basetypes.StringValue `tfsdk:"permission"`
}

// containerResource is the resource implementation.
type containerResource struct {
	client *qnap.Client
}

// NewContainerResource is a helper function to simplify the provider implementation.
func NewContainerResource() resource.Resource {
	return &containerResource{}
}

// Metadata returns the resource type name.
func (r *containerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

// Schema defines the schema for the resource.
func (d *containerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "The last updated timestamp of the container.",
			},
			"status": schema.StringAttribute{
				Required:    true,
				Description: "The state of the container (running, stopped).",
				Validators: []validator.String{
					stringvalidator.OneOf("running", "stopped"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"removeanonvolumes": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to remove anonymous volumes associated with the container.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipaddress": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The ip address assigned to the container incase a networktype bridge is selected.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`), "IP Address must be in a valid format (e.g. 0.0.0.0')."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the container.",
				Validators: []validator.String{
					stringvalidator.OneOf("docker"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the container.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{1,63}$`), "Container name must be between 2 and 64 characters, starts with a letter or number. Valid characters: letters (A-Z, a-z), numbers (0-9), hyphen (-), period (.), underscore (_)"),
				},
			},
			"image": schema.StringAttribute{
				Required:    true,
				Description: "The image of the container.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^(?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)?[a-z0-9]+(?:[._-][a-z0-9]+)*(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$`), "Image name must be in a valid format (e.g. 'nginx:latest', 'myregistry.local:5000/nginx:latest')."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"portbindings": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The host port.",
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"container": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The container port.",
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The protocol used for port binding.",
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"hostip": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The host IP address.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`), "IP Address must be in a valid format (e.g. 0.0.0.0')."),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"restartpolicy": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The name of the restart policy.",
						Validators: []validator.String{
							stringvalidator.OneOf("no", "onFailure", "always", "unlessStopped"),
						},
					},
					"maximumretrycount": schema.Int32Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The maximum number of retries for the restart policy.",
						Validators: []validator.Int32{
							int32validator.Between(0, 1000),
						},
					},
				},
			},
			"autoremove": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to automatically remove the container when it exits.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cmd": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The command to run in the container.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"entrypoint": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The entrypoint for the container.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"tty": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to allocate a pseudo-TTY.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"openstdin": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to open stdin.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"network": schema.StringAttribute{
				Required:    true,
				Description: "The network to connect the container to. Examples of network/networktype compinations: default(the NAT network)/bridge, host/default, bridge/ethx (ethx for the ethernet adaptor you are connecting to when selecting bridge).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"networktype": schema.StringAttribute{
				Required:    true,
				Description: "The type of the network. Examples of network/networktype compinations: default(the NAT network)/bridge, host/default, bridge/ethx (ethx for the ethernet adaptor you are connecting to when selecting bridge).",
				Validators: []validator.String{
					stringvalidator.OneOf("bridge", "host", "none", "ipvlan", "default"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The hostname of the container.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`), "Hostname must be in a valid format (e.g. 'ubuntu' or 'ubuntu-1') long hostname are not valid only short hostname."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The DNS servers for the container.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The environment variables for the container.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The labels for the container.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"volumes": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The type of the volume. Only host and volume types are supported. container is not support as it will not be managed properly with terraform",
							Validators: []validator.String{
								stringvalidator.OneOf("host", "volume"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The name of the volume when using type volume only.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"container": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The container name for the volume.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"source": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The source path for the volume.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^(\/(?:[^\/\0]+\/)*[^\/\0]+)?$`), "Path must be in a valid format (e.g. 'home/user' or '/home/user/file.txt')."),
							},
						},
						"destination": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The destination path for the volume.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^(\/(?:[^\/\0]+\/)*[^\/\0]+)?$`), "Path must be in a valid format (e.g. 'home/user' or '/home/user/file.txt')."),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"permission": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The permission for the volume.",
							Validators: []validator.String{
								stringvalidator.OneOf("readOnly", "writable"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"runtime": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The runtime for the container.",
				Validators: []validator.String{
					stringvalidator.OneOf("runc", "kata-runtime"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"privileged": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to run the container in privileged mode.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"devices": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The name of the device.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"permission": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The permission for the device.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"cpupin": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"cpuids": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The CPU IDs for the container.",
					},
					"type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The type of CPU pinning.",
					},
				},
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"ipaddress": schema.StringAttribute{
							Computed:    true,
							Description: "The ip address assigned to the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"displayname": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"macaddress": schema.StringAttribute{
							Computed:    true,
							Description: "The MAC address of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"gateway": schema.StringAttribute{
							Computed:    true,
							Description: "The gateway of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"networktype": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the network.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"isstaticip": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the network is static IP.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

// Create a new resource.
func (r *containerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var plan ContainerSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Comprehend the plan and create a new container specifictions
	newContainer, diags := ReadStateOrPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new container
	container, err := r.client.CreateContainer(newContainer, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating container",
			"Could not create container, unexpected error: "+err.Error(),
		)
		return
	}

	//Comprehend the new container specs and populate the plan with the new values
	state, diags := WriteState(ctx, container)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// special case for RemoveAnonVolumes as its static to the plan and is used only during destroy
	state.RemoveAnonVolumes = plan.RemoveAnonVolumes
	// special case for network name as it requires side call to qnap to compare the returned name vs the plan name
	state.Network = plan.Network

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state *ContainerSpecModel
	var finalState ContainerSpecModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed order value from QNAP
	containerState, err := r.client.InspectContainer(state.ID.ValueString(), state.Type.ValueString(), &r.client.Token)
	if err != nil {
		// Handle errors, such as resource not found
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An error occurred while reading the resource: "+err.Error(),
		)
		return
	}

	newState, diags := WriteState(ctx, containerState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalState, diags = CompareStates(ctx, state, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// special case for the last updated field and RemoveAnonVolumes
	finalState.RemoveAnonVolumes = state.RemoveAnonVolumes
	// special case for network name as it requires side call to qnap to compare the returned name vs the plan name
	finalState.Network = state.Network

	// Set refreshed state
	diags = resp.State.Set(ctx, finalState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement the update logic
	// QNAP-client-lib is also missing the update logic
	//{"id":"2b4bd83659817408381d948e6599c830498826f70777b610a1af2b68cafa9758","type":"docker","runtime":"runc","name":"bazarr-10","restartPolicy":{"name":"always","maximumRetryCount":0},"networks":[{"id":"c7d58f09271f0c49b2d4e6ee578dc96e3276e88ac1d45d93a6cabca6cc069b4f","name":"bridge","ipAddress":"10.0.3.13","displayName":"Container Network (lxcbr0) (10.0.3.1)","macAddress":"02:42:0a:00:03:0d","gateway":"10.0.3.1","networkType":"default","isStaticIP":false}],"privileged":false,"devices":[],"volumes":[{"type":"volume","name":"volume_1","container":"","source":"/ZFS530_DATA/.qpkg/container-station/docker/volumes/volume_1/_data","destination":"/config","permission":"writable"}],"cpuLimit":1,"memLimit":1073741824,"memReservation":1073741824,"isCpuLimited":true,"isMemoryLimited":true,"isMemoryReservationLimited":true,"isCpuLimitedOld":true,"isMemoryLimitedOld":true,"isMemoryReservationLimitedOld":true,"hasDefaultWebUrlPort":false,"isCpuPinSupported":false,"cpupin":{"type":"shared","cpuIDs":"0"},"extra":{"restart":false}}
	///container-station/api/v3/containers/docker/update | POST
}

func (r *containerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state ContainerSpecModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	_, err := r.client.DeleteContainer(state.ID.ValueString(), state.Type.ValueString(), state.RemoveAnonVolumes.ValueBool(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HashiCups Order",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *containerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*qnap.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *qnap.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

// isNotFound checks if the error.
func isNotFound(mess error) bool {
	var status int
	_, err := fmt.Sscanf(mess.Error(), "status: %d,", &status)
	if err != nil {
		fmt.Println("Error parsing status:", err)
		return false
	}

	// Step 2: Extract the JSON part from the input string
	start := strings.Index(mess.Error(), "body: {")
	if start == -1 {
		fmt.Println("Error: Could not find the body")
		return false
	}
	jsonStr := mess.Error()[start+6:]

	// Step 3: Unmarshal the JSON part
	type Response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	var resp Response
	err = json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return false
	}

	// Output the extracted values
	fmt.Println("Status:", status)
	fmt.Println("Code:", resp.Code)
	fmt.Println("Message:", resp.Message)
	if status == 404 && resp.Code == 1009 && strings.Split(resp.Message, ": ")[1] == "No such container" {
		return true
	}
	return false
}

// ReadStateOrPlan reads the state or plan and returns a new container spec.
func ReadStateOrPlan(ctx context.Context, plan *ContainerSpecModel) (qnap.NewContainerSpec, diag.Diagnostics) {
	// Retrieve values from plan
	diagnostics := diag.Diagnostics{}
	newContainer := qnap.NewContainerSpec{
		Type:        plan.Type.ValueString(),
		Name:        plan.Name.ValueString(),
		Image:       plan.Image.ValueString(),
		AutoRemove:  plan.AutoRemove.ValueBool(),
		Tty:         plan.Tty.ValueBool(),
		OpenStdin:   plan.OpenStdin.ValueBool(),
		Network:     plan.Network.ValueString(),
		NetworkType: plan.NetworkType.ValueString(),
		Hostname:    plan.Hostname.ValueString(),
		Runtime:     plan.Runtime.ValueString(),
		Privileged:  plan.Privileged.ValueBool(),
		IPAddress:   plan.IPAddress.ValueString(),
	}

	// Handle Devices
	if !plan.Devices.IsNull() && !plan.Devices.IsUnknown() {
		var planDevices []DevicesModel
		diags := plan.Devices.ElementsAs(ctx, &planDevices, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewContainerSpec{}, diagnostics
		}
		for _, device := range planDevices {
			// Check if device_key is unknown or null
			if device.Name.IsUnknown() {
				diagnostics.AddWarning("Device key is unknown", "The device_key is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				diagnostics.AddWarning("Device key is null", "The device_key is null, skipping processing.")
				continue
			}

			// Check if device_value is unknown or null
			if device.Name.IsUnknown() {
				diagnostics.AddWarning("Device value is unknown", "The device_value is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				diagnostics.AddWarning("Device value is null", "The device_value is null, skipping processing.")
				continue
			}

			// Safely access the value
			newContainer.Devices = append(newContainer.Devices, qnap.Devices{
				Name:       device.Name.ValueString(),
				Permission: device.Permission.ValueString(),
			})
		}
	}

	// Handle Volumes
	if !plan.Volumes.IsNull() && !plan.Volumes.IsUnknown() {
		var planVolumes []VolumesModel
		diags := plan.Volumes.ElementsAs(ctx, &planVolumes, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewContainerSpec{}, diagnostics
		}

		for _, volume := range planVolumes {
			if volume.Type.IsUnknown() || volume.Type.IsNull() ||
				volume.Name.IsUnknown() || volume.Name.IsNull() ||
				volume.Container.IsUnknown() || volume.Container.IsNull() ||
				volume.Source.IsUnknown() || volume.Source.IsNull() ||
				volume.Destination.IsUnknown() || volume.Destination.IsNull() ||
				volume.Permission.IsUnknown() || volume.Permission.IsNull() {
				diagnostics.AddWarning("Volume attributes are unknown or null", "Skipping processing of a volume because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newContainer.Volumes = append(newContainer.Volumes, qnap.Volumes{
				Type:        volume.Type.ValueString(),
				Name:        volume.Name.ValueString(),
				Container:   volume.Container.ValueString(),
				Source:      volume.Source.ValueString(),
				Destination: volume.Destination.ValueString(),
				Permission:  volume.Permission.ValueString(),
			})
		}
	}

	// Handle PortBindings
	if !plan.PortBindings.IsNull() && !plan.PortBindings.IsUnknown() {
		var planPortBindings []PortBindingsModel
		diags := plan.PortBindings.ElementsAs(ctx, &planPortBindings, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewContainerSpec{}, diagnostics
		}
		for _, portBinding := range planPortBindings {
			if portBinding.Host.IsUnknown() || portBinding.Host.IsNull() ||
				portBinding.Container.IsUnknown() || portBinding.Container.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() {
				diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}
			// Safely access the values
			newContainer.PortBindings = append(newContainer.PortBindings, qnap.PortBindings{
				Host:      portBinding.Host.ValueInt32(),
				Container: portBinding.Container.ValueInt32(),
				Protocol:  portBinding.Protocol.ValueString(),
				HostIP:    portBinding.HostIP.ValueString(),
			})
		}
	}

	// Handle RestartPolicy
	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		var planRestartPolicy RestartPolicyModel
		diags := plan.RestartPolicy.As(ctx, &planRestartPolicy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewContainerSpec{}, diagnostics
		}
		if planRestartPolicy.Name.IsUnknown() || planRestartPolicy.Name.IsNull() ||
			planRestartPolicy.MaximumRetryCount.IsUnknown() || planRestartPolicy.MaximumRetryCount.IsNull() {
			diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
		} else {
			newContainer.RestartPolicy = qnap.RestartPolicy{
				Name:              planRestartPolicy.Name.ValueString(),
				MaximumRetryCount: planRestartPolicy.MaximumRetryCount.ValueInt32(),
			}
		}
	}

	// Handle CPUPIN
	if !plan.Cpupin.IsNull() && !plan.Cpupin.IsUnknown() {
		var planCpupin CpupinModel
		diags := plan.Cpupin.As(ctx, &planCpupin, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewContainerSpec{}, diagnostics
		}
		if planCpupin.CPUIDs.IsUnknown() || planCpupin.CPUIDs.IsNull() ||
			planCpupin.Type.IsUnknown() || planCpupin.Type.IsNull() {
			diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
		} else {
			newContainer.Cpupin = qnap.Cpupin{
				CPUIDs: planCpupin.CPUIDs.ValueString(),
				Type:   planCpupin.Type.ValueString(),
			}
		}
	}

	newContainer.Env = make(map[string]string, len(plan.Env.Elements()))
	_ = plan.Env.ElementsAs(ctx, newContainer.Env, false)

	newContainer.Labels = make(map[string]string, len(plan.Labels.Elements()))
	_ = plan.Labels.ElementsAs(ctx, newContainer.Labels, false)

	for _, item := range plan.Cmd.Elements() {
		newContainer.Cmd = append(newContainer.Cmd, strings.Replace(item.String(), `"`, ``, -1))
		tflog.Debug(ctx, fmt.Sprintf("CMD: %s", strings.Replace(item.String(), `"`, ``, -1)))
	}
	for _, item := range plan.Entrypoint.Elements() {
		newContainer.Entrypoint = append(newContainer.Entrypoint, strings.Replace(item.String(), `"`, ``, -1))
		tflog.Debug(ctx, fmt.Sprintf("Entrypoint: %s", strings.Replace(item.String(), `"`, ``, -1)))
	}
	for _, item := range plan.DNS.Elements() {
		newContainer.DNS = append(newContainer.DNS, strings.Replace(item.String(), `"`, ``, -1))
	}
	return newContainer, diagnostics
}

// WriteState populates the plan with the new values.
func WriteState(ctx context.Context, container *qnap.ContainerInfo) (ContainerSpecModel, diag.Diagnostics) {
	// Map response body to schema and populate Computed attribute values
	plan := ContainerSpecModel{}
	diagnostics := diag.Diagnostics{}

	plan.ID = types.StringValue(container.Data.ID)
	plan.AutoRemove = types.BoolValue(container.Data.AutoRemove)
	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, container.Data.Cmd)
	plan.Tty = types.BoolValue(container.Data.Tty)
	plan.OpenStdin = types.BoolValue(container.Data.OpenStdin)
	plan.Hostname = types.StringValue(container.Data.Hostname)
	plan.Runtime = types.StringValue(container.Data.Runtime)
	plan.Privileged = types.BoolValue(container.Data.Privileged)
	plan.Name = types.StringValue(container.Data.Name)
	plan.Image = types.StringValue(container.Data.Image)
	plan.Type = types.StringValue(container.Data.Type)
	plan.Status = types.StringValue(container.Data.Status)
	plan.NetworkType = types.StringValue(container.Data.Networks[0].NetworkType)

	// Populate Entrypoint attribute
	elements := []attr.Value{}
	for _, item := range container.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	// Populate DNS attribute
	elements = []attr.Value{}
	for _, item := range container.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	// Populate Env attribute
	values := map[string]attr.Value{}
	for envKey, envValue := range container.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	// Populate Labels attribute
	values = map[string]attr.Value{}
	for labelKey, labelValue := range container.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

	// Convert []Networks to basetypes.ListValue
	var networkListElements []attr.Value
	// Define the types for each attribute in the map
	networkAttrTypes := map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"ipaddress":   types.StringType,
		"displayname": types.StringType,
		"macaddress":  types.StringType,
		"gateway":     types.StringType,
		"networktype": types.StringType,
		"isstaticip":  types.BoolType,
	}
	for _, network := range container.Data.Networks {
		// Map the attributes' values
		networkMap := map[string]attr.Value{
			"id":          types.StringValue(network.ID),
			"name":        types.StringValue(network.Name),
			"ipaddress":   types.StringValue(network.IPAddress),
			"displayname": types.StringValue(network.DisplayName),
			"macaddress":  types.StringValue(network.MacAddress),
			"gateway":     types.StringValue(network.Gateway),
			"networktype": types.StringValue(network.NetworkType),
			"isstaticip":  types.BoolValue(network.IsStaticIP),
		}

		networkObject, diags := types.ObjectValue(networkAttrTypes, networkMap)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return ContainerSpecModel{}, diagnostics
		}

		networkListElements = append(networkListElements, networkObject)
	}
	plan.Networks = basetypes.NewListValueMust(types.ObjectType{AttrTypes: networkAttrTypes}, networkListElements)
	if plan.IPAddress.IsNull() || plan.IPAddress.IsUnknown() {
		if len(container.Data.Networks) == 1 && container.Data.Networks[0].IPAddress != "" {
			plan.IPAddress = types.StringValue(container.Data.Networks[0].IPAddress)
		} else {
			plan.IPAddress = types.StringValue("")
		}
	}

	// Convert []Volumes to basetypes.ListValue
	var volumeListElements []attr.Value
	// Define the types for each attribute in the map
	volumeAttrTypes := map[string]attr.Type{
		"type":        types.StringType,
		"name":        types.StringType,
		"container":   types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"permission":  types.StringType,
	}
	for _, volume := range container.Data.Volumes {
		// Map the attributes' values
		volumeMap := map[string]attr.Value{
			"type":        types.StringValue(volume.Type),
			"name":        types.StringValue(volume.Name),
			"container":   types.StringValue(volume.Container),
			"source":      types.StringValue(volume.Source),
			"destination": types.StringValue(volume.Destination),
			"permission":  types.StringValue(volume.Permission),
		}

		volumeObject, diags := types.ObjectValue(volumeAttrTypes, volumeMap)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return ContainerSpecModel{}, diagnostics
		}

		volumeListElements = append(volumeListElements, volumeObject)
	}
	plan.Volumes = basetypes.NewListValueMust(types.ObjectType{AttrTypes: volumeAttrTypes}, volumeListElements)

	// Convert []Devices to basetypes.ListValue
	var deviceListElements []attr.Value
	deviceAttrTypes := map[string]attr.Type{
		"name":       types.StringType,
		"permission": types.StringType,
	}
	for _, device := range container.Data.Devices {
		deviceMap := map[string]attr.Value{
			"name":       types.StringValue(device.Name),
			"permission": types.StringValue(device.Permission),
		}

		deviceObject, diags := types.ObjectValue(deviceAttrTypes, deviceMap)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return ContainerSpecModel{}, diagnostics
		}

		deviceListElements = append(deviceListElements, deviceObject)
	}
	plan.Devices = basetypes.NewListValueMust(types.ObjectType{AttrTypes: deviceAttrTypes}, deviceListElements)

	// Convert []PortBindings to basetypes.ListValue
	var portBindingListElements []attr.Value
	portBindingAttrTypes := map[string]attr.Type{
		"host":      types.Int32Type,
		"container": types.Int32Type,
		"protocol":  types.StringType,
		"hostip":    types.StringType,
	}
	for _, portBinding := range container.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":      types.Int32Value(portBinding.Host),
			"container": types.Int32Value(portBinding.Container),
			"protocol":  types.StringValue(strings.ToLower(portBinding.Protocol)),
			"hostip":    types.StringValue(portBinding.HostIP),
		}

		portBindingObject, diags := types.ObjectValue(portBindingAttrTypes, portBindingMap)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return ContainerSpecModel{}, diagnostics
		}

		portBindingListElements = append(portBindingListElements, portBindingObject)
	}

	plan.PortBindings = basetypes.NewListValueMust(types.ObjectType{AttrTypes: portBindingAttrTypes}, portBindingListElements)

	// Convert RestartPolicy to basetypes.MapValue
	restartPolicyAttrTypes := map[string]attr.Type{
		"name":              types.StringType,
		"maximumretrycount": types.Int32Type,
	}
	restartPolicyMap := map[string]attr.Value{
		"name":              types.StringValue(container.Data.RestartPolicy.Name),
		"maximumretrycount": types.Int32Value(container.Data.RestartPolicy.MaximumRetryCount),
	}

	restartPolicyObject, diags := types.ObjectValue(restartPolicyAttrTypes, restartPolicyMap)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return ContainerSpecModel{}, diagnostics
	}
	plan.RestartPolicy = restartPolicyObject

	// Convert CPUPIN to basetypes.MapValue
	cpupinAttrTypes := map[string]attr.Type{
		"cpuids": types.StringType,
		"type":   types.StringType,
	}
	cpupinMap := map[string]attr.Value{
		"cpuids": types.StringValue(container.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(container.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return ContainerSpecModel{}, diagnostics
	}
	plan.Cpupin = cpupinObject

	// Set last updated field to current time
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	return plan, diagnostics
}

// CompareStates compares the plan and state and returns the state.
func CompareStates(ctx context.Context, plan *ContainerSpecModel, state *ContainerSpecModel) (ContainerSpecModel, diag.Diagnostics) {
	// TODO: Compare only the fields that can be updated
	return *state, diag.Diagnostics{}
}
