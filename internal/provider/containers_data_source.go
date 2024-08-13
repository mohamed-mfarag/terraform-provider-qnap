package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/qnap-client-lib"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &containersDataSource{}
	_ datasource.DataSourceWithConfigure = &containersDataSource{}
)

// containersDataSource is the data source implementation.
type containersDataSource struct {
	client *qnap.Client
}

// containersDataSourceModel maps the data source schema data.
type containersDataSourceModel struct {
	Containers []containersModel `tfsdk:"containers"`
}

// containersModel maps containers schema data.
type containersModel struct {
	ID                    types.String                  `tfsdk:"id"`
	Name                  types.String                  `tfsdk:"name"`
	Type                  types.String                  `tfsdk:"type"`
	Image                 types.String                  `tfsdk:"image"`
	ImageID               types.String                  `tfsdk:"imageid"`
	Status                types.String                  `tfsdk:"status"`
	Project               types.String                  `tfsdk:"project"`
	Runtime               types.String                  `tfsdk:"runtime"`
	MemLimit              types.Int32                   `tfsdk:"memorylimit"`
	CpuLimit              types.Int32                   `tfsdk:"cpulimit"`
	Cpupin                types.Int32                   `tfsdk:"cpupin"`
	UUID                  types.String                  `tfsdk:"uuid"`
	UsedByInternalService types.String                  `tfsdk:"usedbyinternalservice"`
	Privileged            types.Bool                    `tfsdk:"privileged"`
	CPU                   types.Float32                 `tfsdk:"cpu"`
	Memory                types.Float32                 `tfsdk:"memory"`
	TX                    types.Int32                   `tfsdk:"tx"`
	RX                    types.Int32                   `tfsdk:"rx"`
	Read                  types.Int32                   `tfsdk:"read"`
	Write                 types.Int32                   `tfsdk:"write"`
	Created               types.String                  `tfsdk:"created"`
	StartedAt             types.String                  `tfsdk:"startedat"`
	CMD                   types.String                  `tfsdk:"cmd"`
	PortBindings          []containersPortBindingsModel `tfsdk:"portbindings"`
	Networks              []containersNetworksModel     `tfsdk:"networks"`
}

// containersIngredientsModel maps container ingredients data
type containersPortBindingsModel struct {
	Host        types.Int32  `tfsdk:"host"`
	Container   types.Int32  `tfsdk:"container"`
	Protocol    types.String `tfsdk:"protocol"`
	HostIP      types.String `tfsdk:"hostip"`
	ContainerIP types.String `tfsdk:"containerip"`
}

type containersNetworksModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"displayname"`
	IpAddress   types.String `tfsdk:"ipaddress"`
	MacAddress  types.String `tfsdk:"macaddress"`
	Gateway     types.String `tfsdk:"gateway"`
	NetworkType types.String `tfsdk:"networktype"`
	IsStaticIP  types.Bool   `tfsdk:"isstaticip"`
}

// NewContainersDataSource is a helper function to simplify the provider implementation.
func NewContainersDataSource() datasource.DataSource {
	return &containersDataSource{}
}

// Metadata returns the data source type name.
func (d *containersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers"
}

// Schema defines the schema for the data source.
func (d *containersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"containers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"image": schema.StringAttribute{
							Required: true,
						},
						"imageid": schema.StringAttribute{
							Required: true,
						},
						"status": schema.StringAttribute{
							Required: true,
						},
						"project": schema.StringAttribute{
							Required: true,
						},
						"runtime": schema.StringAttribute{
							Required: true,
						},
						"memorylimit": schema.Int32Attribute{
							Required: true,
						},
						"cpulimit": schema.Int32Attribute{
							Required: true,
						},
						"cpupin": schema.Int32Attribute{
							Required: true,
						},
						"uuid": schema.StringAttribute{
							Required: true,
						},
						"usedbyinternalservice": schema.StringAttribute{
							Required: true,
						},
						"privileged": schema.BoolAttribute{
							Required: true,
						},
						"cpu": schema.Float32Attribute{
							Required: true,
						},
						"memory": schema.Float32Attribute{
							Required: true,
						},
						"tx": schema.Int32Attribute{
							Required: true,
						},
						"rx": schema.Int32Attribute{
							Required: true,
						},
						"read": schema.Int32Attribute{
							Required: true,
						},
						"write": schema.Int32Attribute{
							Required: true,
						},
						"created": schema.StringAttribute{
							Required: true,
						},
						"startedat": schema.StringAttribute{
							Required: true,
						},
						"cmd": schema.StringAttribute{
							Required: true,
						},
						"portbindings": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.Int32Attribute{
										Required: true,
									},
									"container": schema.Int32Attribute{
										Required: true,
									},
									"protocol": schema.StringAttribute{
										Required: true,
									},
									"hostip": schema.StringAttribute{
										Required: true,
									},
									"containerip": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"networks": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required: true,
									},
									"name": schema.StringAttribute{
										Required: true,
									},
									"displayname": schema.StringAttribute{
										Required: true,
									},
									"ipaddress": schema.StringAttribute{
										Required: true,
									},
									"macaddress": schema.StringAttribute{
										Required: true,
									},
									"gateway": schema.StringAttribute{
										Required: true,
									},
									"networktype": schema.StringAttribute{
										Required: true,
									},
									"isstaticip": schema.BoolAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *containersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state containersDataSourceModel

	containers, err := d.client.GetContainers()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read QNAP Containers",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, container := range containers {
		containerState := containersModel{
			ID:                    types.StringValue(container.ID),
			Name:                  types.StringValue(container.Name),
			Type:                  types.StringValue(container.Type),
			Image:                 types.StringValue(container.Image),
			ImageID:               types.StringValue(container.ImageID),
			Status:                types.StringValue(container.Status),
			Project:               types.StringValue(container.Project),
			Runtime:               types.StringValue(container.Runtime),
			MemLimit:              types.Int32Value(int32(container.MemLimit)),
			CpuLimit:              types.Int32Value(int32(container.CpuLimit)),
			Cpupin:                types.Int32Value(int32(container.Cpupin)),
			UUID:                  types.StringValue(container.UUID),
			UsedByInternalService: types.StringValue(container.UsedByInternalService),
			Privileged:            types.BoolValue(container.Privileged),
			CPU:                   types.Float32Value(float32(container.CPU)),
			Memory:                types.Float32Value(float32(container.Memory)),
			TX:                    types.Int32Value(int32(container.TX)),
			RX:                    types.Int32Value(int32(container.RX)),
			Read:                  types.Int32Value(int32(container.Read)),
			Write:                 types.Int32Value(int32(container.Write)),
			Created:               types.StringValue(container.Created),
			StartedAt:             types.StringValue(container.StartedAt),
		}
		// Loop on the command array - the array is only configured in the client but configured as string in the provider datasource
		var commands string
		for _, cmd := range container.CMD {
			commands += cmd + " "
		}
		containerState.CMD = types.StringValue(strings.TrimSpace(commands))

		// Process port bindings
		for _, portBinding := range container.PortBindings {
			containerState.PortBindings = append(containerState.PortBindings, containersPortBindingsModel{
				Host:        types.Int32Value(int32(portBinding.Host)),
				Container:   types.Int32Value(int32(portBinding.Container)),
				Protocol:    types.StringValue(portBinding.Protocol),
				HostIP:      types.StringValue(portBinding.HostIP),
				ContainerIP: types.StringValue(portBinding.ContainerIP),
			})
		}

		// Process networks
		for _, network := range container.Networks {
			containerState.Networks = append(containerState.Networks, containersNetworksModel{
				ID:          types.StringValue(network.ID),
				Name:        types.StringValue(network.Name),
				DisplayName: types.StringValue(network.DisplayName),
				IpAddress:   types.StringValue(network.IpAddress),
				MacAddress:  types.StringValue(network.MacAddress),
				Gateway:     types.StringValue(network.Gateway),
				NetworkType: types.StringValue(network.NetworkType),
				IsStaticIP:  types.BoolValue(network.IsStaticIP),
			})
		}

		// Append to state
		state.Containers = append(state.Containers, containerState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *containersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}
