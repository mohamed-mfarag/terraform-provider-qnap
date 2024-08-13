package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/qnap-client-lib"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &volumeResource{}
	_ resource.ResourceWithConfigure = &volumeResource{}
)

type VolumeSpecModel struct {
	ID   basetypes.StringValue `tfsdk:"id"`
	Type basetypes.StringValue `tfsdk:"type"`
	Name basetypes.StringValue `tfsdk:"name"`
}

// volumeResource is the resource implementation.
type volumeResource struct {
	client *qnap.Client
}

// NewVolumeResource is a helper function to simplify the provider implementation.
func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

// Metadata returns the resource type name.
func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the schema for the resource.
func (d *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"removeanonvolumes": schema.BoolAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"image": schema.StringAttribute{
				Required: true,
			},
			"portbindings": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						"volume": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						"protocol": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"hostip": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"volumeip": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"restartpolicy": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"maximumretrycount": schema.Int32Attribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			"autoremove": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"cmd": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"entrypoint": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"tty": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"openstdin": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"network": schema.StringAttribute{
				Required: true,
			},
			"networktype": schema.StringAttribute{
				Required: true,
			},
			"hostname": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"dns": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"volumes": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"name": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"volume": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"source": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"destination": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"permission": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"runtime": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"privileged": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"devices": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"permission": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"cpupin": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"cpuids": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"type": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

// Create a new resource.
func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan VolumeSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newVolume := qnap.NewVolumeSpec{
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
	// Safely iterate over the devices
	if !plan.Devices.IsNull() && !plan.Devices.IsUnknown() {
		var planDevices []Devices
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
			newVolume.Devices = append(newVolume.Devices, qnap.Devices{
				Name:       device.Name.ValueString(),
				Permission: device.Permission.ValueString(),
			})
		}
	}

	// Handle Volumes
	if !plan.Volumes.IsNull() && !plan.Volumes.IsUnknown() {
		var planVolumes []Volumes
		diags := plan.Volumes.ElementsAs(ctx, &planVolumes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, volume := range planVolumes {
			if volume.Type.IsUnknown() || volume.Type.IsNull() ||
				volume.Name.IsUnknown() || volume.Name.IsNull() ||
				volume.Volume.IsUnknown() || volume.Volume.IsNull() ||
				volume.Source.IsUnknown() || volume.Source.IsNull() ||
				volume.Destination.IsUnknown() || volume.Destination.IsNull() ||
				volume.Permission.IsUnknown() || volume.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Volume attributes are unknown or null", "Skipping processing of a volume because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newVolume.Volumes = append(newVolume.Volumes, qnap.Volumes{
				Type:        volume.Type.ValueString(),
				Name:        volume.Name.ValueString(),
				Volume:      volume.Volume.ValueString(),
				Source:      volume.Source.ValueString(),
				Destination: volume.Destination.ValueString(),
				Permission:  volume.Permission.ValueString(),
			})
			// Process the volume
		}
	}

	if !plan.PortBindings.IsNull() && !plan.PortBindings.IsUnknown() {
		var planPortBindings []PortBindings
		diags := plan.PortBindings.ElementsAs(ctx, &planPortBindings, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle PortBindings
		for _, portBinding := range planPortBindings {
			if portBinding.Host.IsUnknown() || portBinding.Host.IsNull() ||
				portBinding.Volume.IsUnknown() || portBinding.Volume.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.VolumeIP.IsUnknown() || portBinding.VolumeIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newVolume.PortBindings = append(newVolume.PortBindings, qnap.PortBindings{
				Host:     portBinding.Host.ValueInt32(),
				Volume:   portBinding.Volume.ValueInt32(),
				Protocol: portBinding.Protocol.ValueString(),
				HostIP:   portBinding.HostIP.ValueString(),
				VolumeIP: portBinding.VolumeIP.ValueString(),
			})
		}
	}

	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		var planRestartPolicy RestartPolicy
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
			newVolume.RestartPolicy = qnap.RestartPolicy{
				Name:              planRestartPolicy.Name.ValueString(),
				MaximumRetryCount: planRestartPolicy.MaximumRetryCount.ValueInt32(),
			}
		}
	}

	if !plan.Cpupin.IsNull() && !plan.Cpupin.IsUnknown() {
		var planCpupin Cpupin
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
			newVolume.Cpupin = qnap.Cpupin{
				CPUIDs: planCpupin.CPUIDs.ValueString(),
				Type:   planCpupin.Type.ValueString(),
			}
		}
	}

	newVolume.Env = make(map[string]string, len(plan.Env.Elements()))
	_ = plan.Env.ElementsAs(ctx, newVolume.Env, false)

	newVolume.Labels = make(map[string]string, len(plan.Labels.Elements()))
	_ = plan.Labels.ElementsAs(ctx, newVolume.Labels, false)

	for _, item := range plan.Cmd.Elements() {
		newVolume.Cmd = append(newVolume.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newVolume.Entrypoint = append(newVolume.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newVolume.DNS = append(newVolume.DNS, item.String())
	}

	// Create new volume
	volume, err := r.client.CreateVolume(newVolume, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating volume",
			"Could not create volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(volume.Data.ID)

	plan.AutoRemove = types.BoolValue(volume.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, volume.Data.Cmd)

	elements := []attr.Value{}
	for _, item := range volume.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range volume.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	plan.Tty = types.BoolValue(volume.Data.Tty)
	plan.OpenStdin = types.BoolValue(volume.Data.OpenStdin)
	plan.Hostname = types.StringValue(volume.Data.Hostname)

	values := map[string]attr.Value{}
	for envKey, envValue := range volume.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range volume.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

	// Convert []Volumes to basetypes.ListValue
	var volumeListElements []attr.Value
	// Define the types for each attribute in the map
	volumeAttrTypes := map[string]attr.Type{
		"type":        types.StringType,
		"name":        types.StringType,
		"volume":      types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"permission":  types.StringType,
	}
	for _, volume := range volume.Data.Volumes {
		// Map the attributes' values
		volumeMap := map[string]attr.Value{
			"type":        types.StringValue(volume.Type),
			"name":        types.StringValue(volume.Name),
			"volume":      types.StringValue(volume.Volume),
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
	for _, device := range volume.Data.Devices {
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
		"host":     types.Int32Type,
		"volume":   types.Int32Type,
		"protocol": types.StringType,
		"hostip":   types.StringType,
		"volumeip": types.StringType,
	}
	for _, portBinding := range volume.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":     types.Int32Value(portBinding.Host),
			"volume":   types.Int32Value(portBinding.Volume),
			"protocol": types.StringValue(portBinding.Protocol),
			"hostip":   types.StringValue(portBinding.HostIP),
			"volumeip": types.StringValue(portBinding.VolumeIP),
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
		"name":              types.StringValue(volume.Data.RestartPolicy.Name),
		"maximumretrycount": types.Int32Value(volume.Data.RestartPolicy.MaximumRetryCount),
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
		"cpuids": types.StringValue(volume.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(volume.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(volume.Data.Runtime)
	plan.Privileged = types.BoolValue(volume.Data.Privileged)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state *VolumeSpecModel
	var mappingDiags diag.Diagnostics
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed order value from QNAP
	volumeState, err := r.client.InspectVolume(state.ID.ValueString(), state.Type.ValueString(), &r.client.Token)
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

	state, mappingDiags = mapFetchedDataToState(*volumeState)
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
func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//The difference between the create and update
	operation := types.StringValue("recreate")
	// Retrieve values from plan
	var plan VolumeSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newVolume := qnap.NewVolumeSpec{
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
		var planDevices []Devices
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
			newVolume.Devices = append(newVolume.Devices, qnap.Devices{
				Name:       device.Name.ValueString(),
				Permission: device.Permission.ValueString(),
			})
		}
	}

	// Handle Volumes
	if !plan.Volumes.IsNull() && !plan.Volumes.IsUnknown() {
		var planVolumes []Volumes
		diags := plan.Volumes.ElementsAs(ctx, &planVolumes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, volume := range planVolumes {
			if volume.Type.IsUnknown() || volume.Type.IsNull() ||
				volume.Name.IsUnknown() || volume.Name.IsNull() ||
				volume.Volume.IsUnknown() || volume.Volume.IsNull() ||
				volume.Source.IsUnknown() || volume.Source.IsNull() ||
				volume.Destination.IsUnknown() || volume.Destination.IsNull() ||
				volume.Permission.IsUnknown() || volume.Permission.IsNull() {
				resp.Diagnostics.AddWarning("Volume attributes are unknown or null", "Skipping processing of a volume because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newVolume.Volumes = append(newVolume.Volumes, qnap.Volumes{
				Type:        volume.Type.ValueString(),
				Name:        volume.Name.ValueString(),
				Volume:      volume.Volume.ValueString(),
				Source:      volume.Source.ValueString(),
				Destination: volume.Destination.ValueString(),
				Permission:  volume.Permission.ValueString(),
			})
			// Process the volume
		}
	}

	if !plan.PortBindings.IsNull() && !plan.PortBindings.IsUnknown() {
		var planPortBindings []PortBindings
		diags := plan.PortBindings.ElementsAs(ctx, &planPortBindings, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle PortBindings
		for _, portBinding := range planPortBindings {
			if portBinding.Host.IsUnknown() || portBinding.Host.IsNull() ||
				portBinding.Volume.IsUnknown() || portBinding.Volume.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.VolumeIP.IsUnknown() || portBinding.VolumeIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newVolume.PortBindings = append(newVolume.PortBindings, qnap.PortBindings{
				Host:     portBinding.Host.ValueInt32(),
				Volume:   portBinding.Volume.ValueInt32(),
				Protocol: portBinding.Protocol.ValueString(),
				HostIP:   portBinding.HostIP.ValueString(),
				VolumeIP: portBinding.VolumeIP.ValueString(),
			})
		}
	}

	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		var planRestartPolicy RestartPolicy
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
			newVolume.RestartPolicy = qnap.RestartPolicy{
				Name:              planRestartPolicy.Name.ValueString(),
				MaximumRetryCount: planRestartPolicy.MaximumRetryCount.ValueInt32(),
			}
		}
	}

	if !plan.Cpupin.IsNull() && !plan.Cpupin.IsUnknown() {
		var planCpupin Cpupin
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
			newVolume.Cpupin = qnap.Cpupin{
				CPUIDs: planCpupin.CPUIDs.ValueString(),
				Type:   planCpupin.Type.ValueString(),
			}
		}
	}

	newVolume.Env = make(map[string]string, len(plan.Env.Elements()))
	_ = plan.Env.ElementsAs(ctx, newVolume.Env, false)

	newVolume.Labels = make(map[string]string, len(plan.Labels.Elements()))
	_ = plan.Labels.ElementsAs(ctx, newVolume.Labels, false)

	for _, item := range plan.Cmd.Elements() {
		newVolume.Cmd = append(newVolume.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newVolume.Entrypoint = append(newVolume.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newVolume.DNS = append(newVolume.DNS, item.String())
	}

	// Create new volume
	volume, err := r.client.CreateVolume(newVolume, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating volume",
			"Could not create volume, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(volume.Data.ID)

	plan.AutoRemove = types.BoolValue(volume.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, volume.Data.Cmd)

	elements := []attr.Value{}
	for _, item := range volume.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range volume.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	plan.Tty = types.BoolValue(volume.Data.Tty)
	plan.OpenStdin = types.BoolValue(volume.Data.OpenStdin)
	plan.Hostname = types.StringValue(volume.Data.Hostname)

	values := map[string]attr.Value{}
	for envKey, envValue := range volume.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range volume.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

	// Convert []Volumes to basetypes.ListValue
	var volumeListElements []attr.Value
	// Define the types for each attribute in the map
	volumeAttrTypes := map[string]attr.Type{
		"type":        types.StringType,
		"name":        types.StringType,
		"volume":      types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"permission":  types.StringType,
	}
	for _, volume := range volume.Data.Volumes {
		// Map the attributes' values
		volumeMap := map[string]attr.Value{
			"type":        types.StringValue(volume.Type),
			"name":        types.StringValue(volume.Name),
			"volume":      types.StringValue(volume.Volume),
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
	for _, device := range volume.Data.Devices {
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
		"host":     types.Int32Type,
		"volume":   types.Int32Type,
		"protocol": types.StringType,
		"hostip":   types.StringType,
		"volumeip": types.StringType,
	}
	for _, portBinding := range volume.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":     types.Int32Value(portBinding.Host),
			"volume":   types.Int32Value(portBinding.Volume),
			"protocol": types.StringValue(portBinding.Protocol),
			"hostip":   types.StringValue(portBinding.HostIP),
			"volumeip": types.StringValue(portBinding.VolumeIP),
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
		"name":              types.StringValue(volume.Data.RestartPolicy.Name),
		"maximumretrycount": types.Int32Value(volume.Data.RestartPolicy.MaximumRetryCount),
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
		"cpuids": types.StringValue(volume.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(volume.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(volume.Data.Runtime)
	plan.Privileged = types.BoolValue(volume.Data.Privileged)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state VolumeSpecModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	_, err := r.client.DeleteVolume(state.ID.ValueString(), state.Type.ValueString(), state.RemoveAnonVolumes.ValueBool(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HashiCups Order",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *volumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
