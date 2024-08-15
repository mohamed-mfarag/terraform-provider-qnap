package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/qnap-client-lib"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

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
	Cpupin            basetypes.ObjectValue `tfsdk:"cpupin"`
	RestartPolicy     basetypes.ObjectValue `tfsdk:"restartpolicy"`
	Cmd               types.List            `tfsdk:"cmd"`
	Entrypoint        basetypes.ListValue   `tfsdk:"entrypoint"`
	DNS               basetypes.ListValue   `tfsdk:"dns"`
	Status            basetypes.StringValue `tfsdk:"status"`
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
	Host        basetypes.Int32Value  `tfsdk:"host"`
	Container   basetypes.Int32Value  `tfsdk:"container"`
	Protocol    basetypes.StringValue `tfsdk:"protocol"`
	HostIP      basetypes.StringValue `tfsdk:"hostip"`
	ContainerIP basetypes.StringValue `tfsdk:"containerip"`
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
			},
			"removeanonvolumes": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to remove anonymous volumes associated with the container.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the container.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the container.",
				Validators: []validator.String{
					stringvalidator.OneOf("docker"),
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
			},
			"portbindings": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The host port.",
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"container": schema.Int32Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The container port.",
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The protocol used for port binding.",
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp"),
							},
						},
						"hostip": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The host IP address.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`), "IP Address must be in a valid format (e.g. 0.0.0.0')."),
							},
						},
						"containerip": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The container IP address.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`), "IP Address must be in a valid format (e.g. 0.0.0.0')."),
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
							stringvalidator.OneOf("no", "on-failure", "always", "unless-stopped"),
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
			},
			"cmd": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The command to run in the container.",
			},
			"entrypoint": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The entrypoint for the container.",
			},
			"tty": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to allocate a pseudo-TTY.",
			},
			"openstdin": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to open stdin.",
			},
			"network": schema.StringAttribute{
				Required:    true,
				Description: "The network to connect the container to.",
			},
			"networktype": schema.StringAttribute{
				Required:    true,
				Description: "The type of the network.",
				Validators: []validator.String{
					stringvalidator.OneOf("bridge", "host", "none", "ipvlan", "default"),
				},
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The hostname of the container.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`), "Hostname must be in a valid format (e.g. 'ubuntu' or 'ubuntu-1') long hostname are not valid only short hostname."),
				},
			},
			"dns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The DNS servers for the container.",
			},
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The environment variables for the container.",
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The labels for the container.",
			},
			"volumes": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The type of the volume.",
							Validators: []validator.String{
								stringvalidator.OneOf("host", "volume", "container"),
							},
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The name of the volume.",
						},
						"container": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The container path for the volume.",
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
						},
						"permission": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The permission for the volume.",
							Validators: []validator.String{
								stringvalidator.OneOf("readOnly", "writable"),
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
			},
			"privileged": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to run the container in privileged mode.",
			},
			"devices": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The name of the device.",
						},
						"permission": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The permission for the device.",
						},
					},
				},
			},
			"cpupin": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
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

	newContainer := qnap.NewContainerSpec{
		Type:        string(plan.Type.ValueString()),
		Name:        string(plan.Name.ValueString()),
		Image:       string(plan.Image.ValueString()),
		AutoRemove:  bool(plan.AutoRemove.ValueBool()),
		Tty:         bool(plan.Tty.ValueBool()),
		OpenStdin:   bool(plan.OpenStdin.ValueBool()),
		Network:     string(plan.Network.ValueString()),
		NetworkType: string(plan.NetworkType.ValueString()),
		Hostname:    string(plan.Hostname.ValueString()),
		Runtime:     string(plan.Runtime.ValueString()),
		Privileged:  bool(plan.Privileged.ValueBool()),
	}

	// Handle Devices
	if !plan.Devices.IsNull() && !plan.Devices.IsUnknown() {
		var planDevices []DevicesModel
		diags := plan.Devices.ElementsAs(ctx, &planDevices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, device := range planDevices {
			// Check if device_key is unknown or null
			if device.Name.IsUnknown() {
				resp.Diagnostics.AddWarning("Device key is unknown", "The device_key is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Device key is null", "The device_key is null, skipping processing.")
				continue
			}

			// Check if device_value is unknown or null
			if device.Name.IsUnknown() {
				resp.Diagnostics.AddWarning("Device value is unknown", "The device_value is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Device value is null", "The device_value is null, skipping processing.")
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, volume := range planVolumes {
			if volume.Type.IsUnknown() || volume.Type.IsNull() ||
				volume.Name.IsUnknown() || volume.Name.IsNull() ||
				volume.Container.IsUnknown() || volume.Container.IsNull() ||
				volume.Source.IsUnknown() || volume.Source.IsNull() ||
				volume.Destination.IsUnknown() || volume.Destination.IsNull() ||
				volume.Permission.IsUnknown() || volume.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Volume attributes are unknown or null", "Skipping processing of a volume because one or more attributes are unknown or null.")
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, portBinding := range planPortBindings {
			if portBinding.Host.IsUnknown() || portBinding.Host.IsNull() ||
				portBinding.Container.IsUnknown() || portBinding.Container.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.ContainerIP.IsUnknown() || portBinding.ContainerIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newContainer.PortBindings = append(newContainer.PortBindings, qnap.PortBindings{
				Host:        portBinding.Host.ValueInt32(),
				Container:   portBinding.Container.ValueInt32(),
				Protocol:    portBinding.Protocol.ValueString(),
				HostIP:      portBinding.HostIP.ValueString(),
				ContainerIP: portBinding.ContainerIP.ValueString(),
			})
		}
	}

	// Handle RestartPolicy
	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		var planRestartPolicy RestartPolicyModel
		diags := plan.RestartPolicy.As(ctx, &planRestartPolicy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if planRestartPolicy.Name.IsUnknown() || planRestartPolicy.Name.IsNull() ||
			planRestartPolicy.MaximumRetryCount.IsUnknown() || planRestartPolicy.MaximumRetryCount.IsNull() {
			resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if planCpupin.CPUIDs.IsUnknown() || planCpupin.CPUIDs.IsNull() ||
			planCpupin.Type.IsUnknown() || planCpupin.Type.IsNull() {
			resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
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
		newContainer.Cmd = append(newContainer.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newContainer.Entrypoint = append(newContainer.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newContainer.DNS = append(newContainer.DNS, item.String())
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

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(container.Data.ID)

	plan.AutoRemove = types.BoolValue(container.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, container.Data.Cmd)

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

	plan.Tty = types.BoolValue(container.Data.Tty)
	plan.OpenStdin = types.BoolValue(container.Data.OpenStdin)
	plan.Hostname = types.StringValue(container.Data.Hostname)

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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		deviceListElements = append(deviceListElements, deviceObject)
	}
	plan.Devices = basetypes.NewListValueMust(types.ObjectType{AttrTypes: deviceAttrTypes}, deviceListElements)

	// Convert []PortBindings to basetypes.ListValue
	var portBindingListElements []attr.Value
	portBindingAttrTypes := map[string]attr.Type{
		"host":        types.Int32Type,
		"container":   types.Int32Type,
		"protocol":    types.StringType,
		"hostip":      types.StringType,
		"containerip": types.StringType,
	}
	for _, portBinding := range container.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":        types.Int32Value(portBinding.Host),
			"container":   types.Int32Value(portBinding.Container),
			"protocol":    types.StringValue(portBinding.Protocol),
			"hostip":      types.StringValue(portBinding.HostIP),
			"containerip": types.StringValue(portBinding.ContainerIP),
		}

		portBindingObject, diags := types.ObjectValue(portBindingAttrTypes, portBindingMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(container.Data.Runtime)
	plan.Privileged = types.BoolValue(container.Data.Privileged)

	//validate the state matches what is expected in the plan
	if plan.Status.ValueString() == "running" && container.Data.Status != "running" {
		_, err = r.client.StartContainer(plan.ID.ValueString(), plan.Type.ValueString(), &r.client.Token)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error change container state to match requested state",
				"Could not start container, unexpected error: "+err.Error(),
			)
			return
		}
	} else if plan.Status.ValueString() == "stopped" && container.Data.Status != "stopped" {
		_, err = r.client.StopContainer(plan.ID.ValueString(), plan.Type.ValueString(), &r.client.Token)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error change container state to match requested state",
				"Could not stop container, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state *ContainerSpecModel
	var mappingDiags diag.Diagnostics
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

	state, mappingDiags = mapFetchedDataToState(*containerState)
	if mappingDiags.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//The difference between the create and update
	operation := types.StringValue("recreate")
	// Retrieve values from plan
	var plan ContainerSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newContainer := qnap.NewContainerSpec{
		Type:        string(plan.Type.ValueString()),
		Name:        string(plan.Name.ValueString()),
		Image:       string(plan.Image.ValueString()),
		AutoRemove:  bool(plan.AutoRemove.ValueBool()),
		Tty:         bool(plan.Tty.ValueBool()),
		OpenStdin:   bool(plan.OpenStdin.ValueBool()),
		Network:     string(plan.Network.ValueString()),
		NetworkType: string(plan.NetworkType.ValueString()),
		Hostname:    string(plan.Hostname.ValueString()),
		Runtime:     string(plan.Runtime.ValueString()),
		Privileged:  bool(plan.Privileged.ValueBool()),
		Operation:   string(operation.ValueString()),
	}
	// Safely iterate over the devices
	if !plan.Devices.IsNull() && !plan.Devices.IsUnknown() {
		var planDevices []DevicesModel
		diags := plan.Devices.ElementsAs(ctx, &planDevices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, device := range planDevices {
			// Check if device_key is unknown or null
			if device.Name.IsUnknown() {
				resp.Diagnostics.AddWarning("Device key is unknown", "The device_key is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Device key is null", "The device_key is null, skipping processing.")
				continue
			}

			// Check if device_value is unknown or null
			if device.Name.IsUnknown() {
				resp.Diagnostics.AddWarning("Device value is unknown", "The device_value is unknown, skipping processing.")
				continue
			}

			if device.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Device value is null", "The device_value is null, skipping processing.")
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, volume := range planVolumes {
			if volume.Type.IsUnknown() || volume.Type.IsNull() ||
				volume.Name.IsUnknown() || volume.Name.IsNull() ||
				volume.Container.IsUnknown() || volume.Container.IsNull() ||
				volume.Source.IsUnknown() || volume.Source.IsNull() ||
				volume.Destination.IsUnknown() || volume.Destination.IsNull() ||
				volume.Permission.IsUnknown() || volume.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Volume attributes are unknown or null", "Skipping processing of a volume because one or more attributes are unknown or null.")
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
			// Process the volume
		}
	}

	if !plan.PortBindings.IsNull() && !plan.PortBindings.IsUnknown() {
		var planPortBindings []PortBindingsModel
		diags := plan.PortBindings.ElementsAs(ctx, &planPortBindings, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle PortBindings
		for _, portBinding := range planPortBindings {
			if portBinding.Host.IsUnknown() || portBinding.Host.IsNull() ||
				portBinding.Container.IsUnknown() || portBinding.Container.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.ContainerIP.IsUnknown() || portBinding.ContainerIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newContainer.PortBindings = append(newContainer.PortBindings, qnap.PortBindings{
				Host:        portBinding.Host.ValueInt32(),
				Container:   portBinding.Container.ValueInt32(),
				Protocol:    portBinding.Protocol.ValueString(),
				HostIP:      portBinding.HostIP.ValueString(),
				ContainerIP: portBinding.ContainerIP.ValueString(),
			})
		}
	}

	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		var planRestartPolicy RestartPolicyModel
		diags := plan.RestartPolicy.As(ctx, &planRestartPolicy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle RestartPolicy
		if planRestartPolicy.Name.IsUnknown() || planRestartPolicy.Name.IsNull() ||
			planRestartPolicy.MaximumRetryCount.IsUnknown() || planRestartPolicy.MaximumRetryCount.IsNull() {
			resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
		} else {
			newContainer.RestartPolicy = qnap.RestartPolicy{
				Name:              planRestartPolicy.Name.ValueString(),
				MaximumRetryCount: planRestartPolicy.MaximumRetryCount.ValueInt32(),
			}
		}
	}

	if !plan.Cpupin.IsNull() && !plan.Cpupin.IsUnknown() {
		var planCpupin CpupinModel
		diags := plan.Cpupin.As(ctx, &planCpupin, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle CPUPIN
		if planCpupin.CPUIDs.IsUnknown() || planCpupin.CPUIDs.IsNull() ||
			planCpupin.Type.IsUnknown() || planCpupin.Type.IsNull() {
			resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
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
		newContainer.Cmd = append(newContainer.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newContainer.Entrypoint = append(newContainer.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newContainer.DNS = append(newContainer.DNS, item.String())
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

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(container.Data.ID)

	plan.AutoRemove = types.BoolValue(container.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, container.Data.Cmd)

	elements := []attr.Value{}
	for _, item := range container.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range container.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	plan.Tty = types.BoolValue(container.Data.Tty)
	plan.OpenStdin = types.BoolValue(container.Data.OpenStdin)
	plan.Hostname = types.StringValue(container.Data.Hostname)

	values := map[string]attr.Value{}
	for envKey, envValue := range container.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range container.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
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
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		deviceListElements = append(deviceListElements, deviceObject)
	}
	plan.Devices = basetypes.NewListValueMust(types.ObjectType{AttrTypes: deviceAttrTypes}, deviceListElements)

	// Convert []PortBindings to basetypes.ListValue
	var portBindingListElements []attr.Value
	portBindingAttrTypes := map[string]attr.Type{
		"host":        types.Int32Type,
		"container":   types.Int32Type,
		"protocol":    types.StringType,
		"hostip":      types.StringType,
		"containerip": types.StringType,
	}
	for _, portBinding := range container.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":        types.Int32Value(portBinding.Host),
			"container":   types.Int32Value(portBinding.Container),
			"protocol":    types.StringValue(portBinding.Protocol),
			"hostip":      types.StringValue(portBinding.HostIP),
			"containerip": types.StringValue(portBinding.ContainerIP),
		}

		portBindingObject, diags := types.ObjectValue(portBindingAttrTypes, portBindingMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(container.Data.Runtime)
	plan.Privileged = types.BoolValue(container.Data.Privileged)

	//validate the state matches what is expected in the plan
	if plan.Status.ValueString() == "running" && container.Data.Status != "running" {
		_, err = r.client.StartContainer(plan.ID.ValueString(), plan.Type.ValueString(), &r.client.Token)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error change container state to match requested state",
				"Could not start container, unexpected error: "+err.Error(),
			)
			return
		}
	} else if plan.Status.ValueString() == "stopped" && container.Data.Status != "stopped" {
		_, err = r.client.StopContainer(plan.ID.ValueString(), plan.Type.ValueString(), &r.client.Token)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error change container state to match requested state",
				"Could not stop container, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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

func mapFetchedDataToState(container qnap.ContainerInfo) (*ContainerSpecModel, diag.Diagnostics) {

	var plan ContainerSpecModel
	var tempDiag diag.Diagnostics

	// Map response body to schema and populate Computed attribute values

	plan.ID = types.StringValue(container.Data.ID)
	plan.Tty = types.BoolValue(container.Data.Tty)
	plan.OpenStdin = types.BoolValue(container.Data.OpenStdin)
	plan.Hostname = types.StringValue(container.Data.Hostname)
	plan.Runtime = types.StringValue(container.Data.Runtime)
	plan.Privileged = types.BoolValue(container.Data.Privileged)

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
	tempDiag.Append(diags...)
	if tempDiag.HasError() {
		return nil, tempDiag
	}
	plan.RestartPolicy = restartPolicyObject

	// Convert CPUPIN to basetypes.MapValue
	cpupinAttrTypes := map[string]attr.Type{
		"name":              types.StringType,
		"maximumretrycount": types.Int32Type,
	}
	cpupinMap := map[string]attr.Value{
		"cpuids": types.StringValue(container.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(container.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	tempDiag.Append(diags...)
	if tempDiag.HasError() {
		return nil, tempDiag
	}
	plan.Cpupin = cpupinObject

	plan.AutoRemove = types.BoolValue(container.Data.AutoRemove)

	elements := []attr.Value{}
	for _, item := range container.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range container.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	values := map[string]attr.Value{}
	for envKey, envValue := range container.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range container.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

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
		tempDiag.Append(diags...)
		if tempDiag.HasError() {
			return nil, tempDiag
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
		tempDiag.Append(diags...)
		if tempDiag.HasError() {
			return nil, tempDiag
		}

		deviceListElements = append(deviceListElements, deviceObject)
	}
	plan.Devices = basetypes.NewListValueMust(types.ObjectType{AttrTypes: deviceAttrTypes}, deviceListElements)

	// Convert []PortBindings to basetypes.ListValue
	var portBindingListElements []attr.Value
	portBindingAttrTypes := map[string]attr.Type{
		"host":        types.Int32Type,
		"container":   types.Int32Type,
		"protocol":    types.StringType,
		"hostip":      types.StringType,
		"containerip": types.StringType,
	}
	for _, portBinding := range container.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":        types.Int32Value(portBinding.Host),
			"container":   types.Int32Value(portBinding.Container),
			"protocol":    types.StringValue(portBinding.Protocol),
			"hostip":      types.StringValue(portBinding.HostIP),
			"containerip": types.StringValue(portBinding.ContainerIP),
		}

		portBindingObject, diags := types.ObjectValue(portBindingAttrTypes, portBindingMap)
		tempDiag.Append(diags...)
		if tempDiag.HasError() {
			return nil, tempDiag
		}

		portBindingListElements = append(portBindingListElements, portBindingObject)
	}

	// Convert the Cmd field from []string to types.ListValue
	newCmd := func() types.List {
		var listElements []attr.Value
		for _, str := range container.Data.Cmd {
			listElements = append(listElements, types.StringValue(str))
		}
		listValue, err := basetypes.NewListValue(types.StringType, listElements)
		if err != nil {
			fmt.Println(err)
		}
		return listValue
	}()
	fmt.Println(newCmd.String())
	plan.Cmd = newCmd
	plan.PortBindings = basetypes.NewListValueMust(types.ObjectType{AttrTypes: portBindingAttrTypes}, portBindingListElements)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return &plan, tempDiag
}

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
