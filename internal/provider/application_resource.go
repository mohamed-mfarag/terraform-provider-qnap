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
	_ resource.Resource              = &appResource{}
	_ resource.ResourceWithConfigure = &appResource{}
)

type AppSpecModel struct {
	ID   basetypes.StringValue `tfsdk:"id"`
	Type basetypes.StringValue `tfsdk:"type"`
	Name basetypes.StringValue `tfsdk:"name"`
}

// appResource is the resource implementation.
type appResource struct {
	client *qnap.Client
}

// NewAppResource is a helper function to simplify the provider implementation.
func NewAppResource() resource.Resource {
	return &appResource{}
}

// Metadata returns the resource type name.
func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

// Schema defines the schema for the resource.
func (d *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"removeanonapps": schema.BoolAttribute{
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
						"app": schema.Int32Attribute{
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
						"appip": schema.StringAttribute{
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
			"apps": schema.ListNestedAttribute{
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
						"app": schema.StringAttribute{
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
func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan AppSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newApp := qnap.NewAppSpec{
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
			newApp.Devices = append(newApp.Devices, qnap.Devices{
				Name:       device.Name.ValueString(),
				Permission: device.Permission.ValueString(),
			})
		}
	}

	// Handle Apps
	if !plan.Apps.IsNull() && !plan.Apps.IsUnknown() {
		var planApps []Apps
		diags := plan.Apps.ElementsAs(ctx, &planApps, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, app := range planApps {
			if app.Type.IsUnknown() || app.Type.IsNull() ||
				app.Name.IsUnknown() || app.Name.IsNull() ||
				app.App.IsUnknown() || app.App.IsNull() ||
				app.Source.IsUnknown() || app.Source.IsNull() ||
				app.Destination.IsUnknown() || app.Destination.IsNull() ||
				app.Permission.IsUnknown() || app.Permission.IsNull() {
				resp.Diagnostics.AddWarning("App attributes are unknown or null", "Skipping processing of a app because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newApp.Apps = append(newApp.Apps, qnap.Apps{
				Type:        app.Type.ValueString(),
				Name:        app.Name.ValueString(),
				App:         app.App.ValueString(),
				Source:      app.Source.ValueString(),
				Destination: app.Destination.ValueString(),
				Permission:  app.Permission.ValueString(),
			})
			// Process the app
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
				portBinding.App.IsUnknown() || portBinding.App.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.AppIP.IsUnknown() || portBinding.AppIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newApp.PortBindings = append(newApp.PortBindings, qnap.PortBindings{
				Host:     portBinding.Host.ValueInt32(),
				App:      portBinding.App.ValueInt32(),
				Protocol: portBinding.Protocol.ValueString(),
				HostIP:   portBinding.HostIP.ValueString(),
				AppIP:    portBinding.AppIP.ValueString(),
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
			newApp.RestartPolicy = qnap.RestartPolicy{
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
			newApp.Cpupin = qnap.Cpupin{
				CPUIDs: planCpupin.CPUIDs.ValueString(),
				Type:   planCpupin.Type.ValueString(),
			}
		}
	}

	newApp.Env = make(map[string]string, len(plan.Env.Elements()))
	_ = plan.Env.ElementsAs(ctx, newApp.Env, false)

	newApp.Labels = make(map[string]string, len(plan.Labels.Elements()))
	_ = plan.Labels.ElementsAs(ctx, newApp.Labels, false)

	for _, item := range plan.Cmd.Elements() {
		newApp.Cmd = append(newApp.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newApp.Entrypoint = append(newApp.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newApp.DNS = append(newApp.DNS, item.String())
	}

	// Create new app
	app, err := r.client.CreateApp(newApp, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(app.Data.ID)

	plan.AutoRemove = types.BoolValue(app.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, app.Data.Cmd)

	elements := []attr.Value{}
	for _, item := range app.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range app.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	plan.Tty = types.BoolValue(app.Data.Tty)
	plan.OpenStdin = types.BoolValue(app.Data.OpenStdin)
	plan.Hostname = types.StringValue(app.Data.Hostname)

	values := map[string]attr.Value{}
	for envKey, envValue := range app.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range app.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

	// Convert []Apps to basetypes.ListValue
	var appListElements []attr.Value
	// Define the types for each attribute in the map
	appAttrTypes := map[string]attr.Type{
		"type":        types.StringType,
		"name":        types.StringType,
		"app":         types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"permission":  types.StringType,
	}
	for _, app := range app.Data.Apps {
		// Map the attributes' values
		appMap := map[string]attr.Value{
			"type":        types.StringValue(app.Type),
			"name":        types.StringValue(app.Name),
			"app":         types.StringValue(app.App),
			"source":      types.StringValue(app.Source),
			"destination": types.StringValue(app.Destination),
			"permission":  types.StringValue(app.Permission),
		}

		appObject, diags := types.ObjectValue(appAttrTypes, appMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		appListElements = append(appListElements, appObject)
	}
	plan.Apps = basetypes.NewListValueMust(types.ObjectType{AttrTypes: appAttrTypes}, appListElements)

	// Convert []Devices to basetypes.ListValue
	var deviceListElements []attr.Value
	deviceAttrTypes := map[string]attr.Type{
		"name":       types.StringType,
		"permission": types.StringType,
	}
	for _, device := range app.Data.Devices {
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
		"app":      types.Int32Type,
		"protocol": types.StringType,
		"hostip":   types.StringType,
		"appip":    types.StringType,
	}
	for _, portBinding := range app.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":     types.Int32Value(portBinding.Host),
			"app":      types.Int32Value(portBinding.App),
			"protocol": types.StringValue(portBinding.Protocol),
			"hostip":   types.StringValue(portBinding.HostIP),
			"appip":    types.StringValue(portBinding.AppIP),
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
		"name":              types.StringValue(app.Data.RestartPolicy.Name),
		"maximumretrycount": types.Int32Value(app.Data.RestartPolicy.MaximumRetryCount),
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
		"cpuids": types.StringValue(app.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(app.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(app.Data.Runtime)
	plan.Privileged = types.BoolValue(app.Data.Privileged)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state *AppSpecModel
	var mappingDiags diag.Diagnostics
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed order value from QNAP
	appState, err := r.client.InspectApp(state.ID.ValueString(), state.Type.ValueString(), &r.client.Token)
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

	state, mappingDiags = mapFetchedDataToState(*appState)
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
func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//The difference between the create and update
	operation := types.StringValue("recreate")
	// Retrieve values from plan
	var plan AppSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newApp := qnap.NewAppSpec{
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
			newApp.Devices = append(newApp.Devices, qnap.Devices{
				Name:       device.Name.ValueString(),
				Permission: device.Permission.ValueString(),
			})
		}
	}

	// Handle Apps
	if !plan.Apps.IsNull() && !plan.Apps.IsUnknown() {
		var planApps []Apps
		diags := plan.Apps.ElementsAs(ctx, &planApps, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, app := range planApps {
			if app.Type.IsUnknown() || app.Type.IsNull() ||
				app.Name.IsUnknown() || app.Name.IsNull() ||
				app.App.IsUnknown() || app.App.IsNull() ||
				app.Source.IsUnknown() || app.Source.IsNull() ||
				app.Destination.IsUnknown() || app.Destination.IsNull() ||
				app.Permission.IsUnknown() || app.Permission.IsNull() {
				resp.Diagnostics.AddWarning("App attributes are unknown or null", "Skipping processing of a app because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newApp.Apps = append(newApp.Apps, qnap.Apps{
				Type:        app.Type.ValueString(),
				Name:        app.Name.ValueString(),
				App:         app.App.ValueString(),
				Source:      app.Source.ValueString(),
				Destination: app.Destination.ValueString(),
				Permission:  app.Permission.ValueString(),
			})
			// Process the app
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
				portBinding.App.IsUnknown() || portBinding.App.IsNull() ||
				portBinding.Protocol.IsUnknown() || portBinding.Protocol.IsNull() ||
				portBinding.HostIP.IsUnknown() || portBinding.HostIP.IsNull() ||
				portBinding.AppIP.IsUnknown() || portBinding.AppIP.IsNull() {
				resp.Diagnostics.AddWarning("Port binding attributes are unknown or null", "Skipping processing of a port binding because one or more attributes are unknown or null.")
				continue
			}

			// Safely access the values
			newApp.PortBindings = append(newApp.PortBindings, qnap.PortBindings{
				Host:     portBinding.Host.ValueInt32(),
				App:      portBinding.App.ValueInt32(),
				Protocol: portBinding.Protocol.ValueString(),
				HostIP:   portBinding.HostIP.ValueString(),
				AppIP:    portBinding.AppIP.ValueString(),
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
			newApp.RestartPolicy = qnap.RestartPolicy{
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
			newApp.Cpupin = qnap.Cpupin{
				CPUIDs: planCpupin.CPUIDs.ValueString(),
				Type:   planCpupin.Type.ValueString(),
			}
		}
	}

	newApp.Env = make(map[string]string, len(plan.Env.Elements()))
	_ = plan.Env.ElementsAs(ctx, newApp.Env, false)

	newApp.Labels = make(map[string]string, len(plan.Labels.Elements()))
	_ = plan.Labels.ElementsAs(ctx, newApp.Labels, false)

	for _, item := range plan.Cmd.Elements() {
		newApp.Cmd = append(newApp.Cmd, item.String())
	}
	for _, item := range plan.Entrypoint.Elements() {
		newApp.Entrypoint = append(newApp.Entrypoint, item.String())
	}
	for _, item := range plan.DNS.Elements() {
		newApp.DNS = append(newApp.DNS, item.String())
	}

	// Create new app
	app, err := r.client.CreateApp(newApp, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(app.Data.ID)

	plan.AutoRemove = types.BoolValue(app.Data.AutoRemove)

	plan.Cmd, _ = types.ListValueFrom(ctx, types.StringType, app.Data.Cmd)

	elements := []attr.Value{}
	for _, item := range app.Data.Entrypoint {
		elements = append(elements, types.StringValue(item))
	}
	plan.Entrypoint, _ = types.ListValue(types.StringType, elements)

	elements = []attr.Value{}
	for _, item := range app.Data.DNS {
		elements = append(elements, types.StringValue(item))
	}
	plan.DNS, _ = types.ListValue(types.StringType, elements)

	plan.Tty = types.BoolValue(app.Data.Tty)
	plan.OpenStdin = types.BoolValue(app.Data.OpenStdin)
	plan.Hostname = types.StringValue(app.Data.Hostname)

	values := map[string]attr.Value{}
	for envKey, envValue := range app.Data.Env {
		values[envKey] = types.StringValue(envValue)
	}
	plan.Env, _ = types.MapValue(types.StringType, values)

	values = map[string]attr.Value{}
	for labelKey, labelValue := range app.Data.Labels {
		values[labelKey] = types.StringValue(labelValue)
	}
	plan.Labels, _ = types.MapValue(types.StringType, values)

	// Convert []Apps to basetypes.ListValue
	var appListElements []attr.Value
	// Define the types for each attribute in the map
	appAttrTypes := map[string]attr.Type{
		"type":        types.StringType,
		"name":        types.StringType,
		"app":         types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"permission":  types.StringType,
	}
	for _, app := range app.Data.Apps {
		// Map the attributes' values
		appMap := map[string]attr.Value{
			"type":        types.StringValue(app.Type),
			"name":        types.StringValue(app.Name),
			"app":         types.StringValue(app.App),
			"source":      types.StringValue(app.Source),
			"destination": types.StringValue(app.Destination),
			"permission":  types.StringValue(app.Permission),
		}

		appObject, diags := types.ObjectValue(appAttrTypes, appMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		appListElements = append(appListElements, appObject)
	}
	plan.Apps = basetypes.NewListValueMust(types.ObjectType{AttrTypes: appAttrTypes}, appListElements)

	// Convert []Devices to basetypes.ListValue
	var deviceListElements []attr.Value
	deviceAttrTypes := map[string]attr.Type{
		"name":       types.StringType,
		"permission": types.StringType,
	}
	for _, device := range app.Data.Devices {
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
		"app":      types.Int32Type,
		"protocol": types.StringType,
		"hostip":   types.StringType,
		"appip":    types.StringType,
	}
	for _, portBinding := range app.Data.PortBindings {
		portBindingMap := map[string]attr.Value{
			"host":     types.Int32Value(portBinding.Host),
			"app":      types.Int32Value(portBinding.App),
			"protocol": types.StringValue(portBinding.Protocol),
			"hostip":   types.StringValue(portBinding.HostIP),
			"appip":    types.StringValue(portBinding.AppIP),
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
		"name":              types.StringValue(app.Data.RestartPolicy.Name),
		"maximumretrycount": types.Int32Value(app.Data.RestartPolicy.MaximumRetryCount),
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
		"cpuids": types.StringValue(app.Data.Cpupin.CPUIDs),
		"type":   types.StringValue(app.Data.Cpupin.Type),
	}

	cpupinObject, diags := types.ObjectValue(cpupinAttrTypes, cpupinMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Cpupin = cpupinObject

	plan.Runtime = types.StringValue(app.Data.Runtime)
	plan.Privileged = types.BoolValue(app.Data.Privileged)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state AppSpecModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	_, err := r.client.DeleteApp(state.ID.ValueString(), state.Type.ValueString(), state.RemoveAnonApps.ValueBool(), &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HashiCups Order",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *appResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
