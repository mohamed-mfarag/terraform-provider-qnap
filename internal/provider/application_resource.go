package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/qnap-client-lib"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gopkg.in/yaml.v2"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &appResource{}
	_ resource.ResourceWithConfigure = &appResource{}
)

type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
	Volumes  map[string]Volume  `yaml:"volumes,omitempty"`
	Networks map[string]Network `yaml:"networks,omitempty"`
}

type Service struct {
	Image       string            `yaml:"image,omitempty"`
	Build       BuildConfig       `yaml:"build,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
}

type BuildConfig struct {
	Context    string            `yaml:"context,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
}

type Volume struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
}

type Network struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
}

type AppSpecModel struct {
	LastUpdated       basetypes.StringValue `tfsdk:"last_updated"`
	Name              basetypes.StringValue `tfsdk:"name"`
	Yml               basetypes.StringValue `tfsdk:"yml"`
	DefaultURL        basetypes.ObjectValue `tfsdk:"default_url"`
	Containers        basetypes.ListValue   `tfsdk:"containers"`
	CPULimit          basetypes.Int32Value  `tfsdk:"cpu_limit"`
	MemLimit          basetypes.Int32Value  `tfsdk:"mem_limit"`
	MemReservation    basetypes.Int32Value  `tfsdk:"mem_reservation"`
	RemoveAnonVolumes basetypes.BoolValue   `tfsdk:"removeanonvolumes"`
}
type ContainersModel struct {
	ID   basetypes.StringValue `tfsdk:"id"`
	Name basetypes.StringValue `tfsdk:"name"`
}
type DefaultURLModel struct {
	Port    basetypes.Int32Value  `tfsdk:"port"`
	Service basetypes.StringValue `tfsdk:"service"`
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the application.",
			},
			"yml": schema.StringAttribute{
				Required:    true,
				Description: "The YAML configuration for the application.",
			},
			"removeanonvolumes": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to remove anonymous volumes when the application is removed.",
			},
			"containers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the container.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the container.",
						},
					},
				},
				Description: "The list of containers in the application.",
			},
			"default_url": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"port": schema.Int32Attribute{
						Optional:    true,
						Description: "The port number for the default URL.",
					},
					"service": schema.StringAttribute{
						Optional:    true,
						Description: "The service name for the default URL.",
					},
				},
				Description: "The default URL for the application.",
			},
			"cpu_limit": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The CPU limit for the application.",
			},
			"mem_limit": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory limit for the application.",
			},
			"mem_reservation": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory reservation for the application.",
			},
			"last_updated": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The last updated timestamp of the application.",
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

	// Validate and convert YAML to JSON
	jsonString, err := validateYAML(plan.Yml.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error validating and converting YAML to JSON",
			err.Error(),
		)
		return
	}

	newApp := qnap.NewAppReqModel{
		Name:           plan.Name.ValueString(),
		Yml:            jsonString,
		CPULimit:       plan.CPULimit.ValueInt32(),
		MemLimit:       plan.MemLimit.ValueInt32(),
		MemReservation: plan.MemReservation.ValueInt32(),
	}

	if !plan.DefaultURL.IsNull() && !plan.DefaultURL.IsUnknown() {
		var default_url DefaultURLModel
		diags := plan.DefaultURL.As(ctx, &default_url, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle RestartPolicy
		if default_url.Service.IsUnknown() || default_url.Service.IsNull() ||
			default_url.Port.IsUnknown() || default_url.Port.IsNull() {
			resp.Diagnostics.AddWarning("missing default url attributes", "Default URL is present however some attributes are missing, ensure service and port are present within the default url attribute.")
		} else {
			newApp.DefaultURL = qnap.NewAppReqDefaultURLModel{
				Service: default_url.Service.ValueString(),
				Port:    default_url.Port.ValueInt32(),
			}
		}
	}

	// Create new app
	app, err := r.client.CreateApplication(newApp, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values

	plan.CPULimit = types.Int32Value(app.Data.CPULimit)
	plan.MemReservation = types.Int32Value(app.Data.MemReservation)
	plan.MemLimit = types.Int32Value(app.Data.MemLimit)

	// Convert []containers to basetypes.ListValue
	var containerListElements []attr.Value
	// Define the types for each attribute in the map
	containerAttrTypes := map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}
	for _, container := range app.Data.Containers {
		// Map the attributes' values
		containerMap := map[string]attr.Value{
			"name": types.StringValue(container.Name),
			"id":   types.StringValue(container.ID),
		}

		containerObject, diags := types.ObjectValue(containerAttrTypes, containerMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		containerListElements = append(containerListElements, containerObject)
	}
	plan.Containers = basetypes.NewListValueMust(types.ObjectType{AttrTypes: containerAttrTypes}, containerListElements)

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed order value from QNAP
	appState, err := r.client.InspectApplication(state.Name.ValueString(), &r.client.Token)
	if err != nil {
		//Handle errors, such as resource not found
		if isAppNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Resource",
			"An error occurred while reading the resource: "+err.Error(),
		)
		return
	}

	// Map response attributes to state attributes
	state.Yml = types.StringValue(appState.Data.Yml)
	state.CPULimit = types.Int32Value(appState.Data.CPULimit)
	state.MemLimit = types.Int32Value(appState.Data.MemLimit)
	state.MemReservation = types.Int32Value(appState.Data.MemReservation)

	// Convert []containers to basetypes.ListValue
	var containerListElements []attr.Value
	// Define the types for each attribute in the map
	containerAttrTypes := map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}
	for _, container := range appState.Data.Containers {
		// Map the attributes' values
		containerMap := map[string]attr.Value{
			"name": types.StringValue(container.Name),
			"id":   types.StringValue(container.ID),
		}

		containerObject, diags := types.ObjectValue(containerAttrTypes, containerMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		containerListElements = append(containerListElements, containerObject)
	}
	state.Containers = basetypes.NewListValueMust(types.ObjectType{AttrTypes: containerAttrTypes}, containerListElements)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

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

	// Validate and convert YAML to JSON
	jsonString, err := validateYAML(plan.Yml.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error validating and converting YAML to JSON",
			err.Error(),
		)
		return
	}

	newApp := qnap.NewAppReqModel{
		Name:           plan.Name.ValueString(),
		Yml:            jsonString,
		CPULimit:       plan.CPULimit.ValueInt32(),
		MemLimit:       plan.MemLimit.ValueInt32(),
		MemReservation: plan.MemReservation.ValueInt32(),
		Operation:      operation.ValueString(),
	}

	if !plan.DefaultURL.IsNull() && !plan.DefaultURL.IsUnknown() {
		var default_url DefaultURLModel
		diags := plan.DefaultURL.As(ctx, &default_url, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Handle RestartPolicy
		if default_url.Service.IsUnknown() || default_url.Service.IsNull() ||
			default_url.Port.IsUnknown() || default_url.Port.IsNull() {
			resp.Diagnostics.AddWarning("missing default url attributes", "Default URL is present however some attributes are missing, ensure service and port are present within the default url attribute.")
		} else {
			newApp.DefaultURL = qnap.NewAppReqDefaultURLModel{
				Service: default_url.Service.ValueString(),
				Port:    default_url.Port.ValueInt32(),
			}
		}
	}

	// Create new app
	app, err := r.client.CreateApplication(newApp, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values

	plan.CPULimit = types.Int32Value(app.Data.CPULimit)
	plan.MemReservation = types.Int32Value(app.Data.MemReservation)
	plan.MemLimit = types.Int32Value(app.Data.MemLimit)

	// Convert []containers to basetypes.ListValue
	var containerListElements []attr.Value
	// Define the types for each attribute in the map
	containerAttrTypes := map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}
	for _, container := range app.Data.Containers {
		// Map the attributes' values
		containerMap := map[string]attr.Value{
			"name": types.StringValue(container.Name),
			"id":   types.StringValue(container.ID),
		}

		containerObject, diags := types.ObjectValue(containerAttrTypes, containerMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		containerListElements = append(containerListElements, containerObject)
	}
	plan.Containers = basetypes.NewListValueMust(types.ObjectType{AttrTypes: containerAttrTypes}, containerListElements)

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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
	_, err := r.client.DeleteApplication(state.Name.ValueString(), state.RemoveAnonVolumes.ValueBool(), &r.client.Token)
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

func validateYAML(yamlData string) (string, error) {
	var compose ComposeFile
	err := yaml.Unmarshal([]byte(yamlData), &compose)
	if err != nil {
		return "", fmt.Errorf("invalid YAML: %w", err)
	}

	if compose.Version == "" {
		return "", fmt.Errorf("missing required field: version")
	}
	if len(compose.Services) == 0 {
		return "", fmt.Errorf("no services defined")
	}

	for serviceName, service := range compose.Services {
		if service.Image == "" && service.Build.Context == "" {
			return "", fmt.Errorf("service '%s' must have either an image or a build context", serviceName)
		}
	}

	ValidatedYamlData, err := yaml.Marshal(compose)
	if err != nil {
		return "", fmt.Errorf("error converting to YAMML: %w", err)
	}

	return string(ValidatedYamlData), nil
}

func isAppNotFound(mess error) bool {
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
	if status == 404 && resp.Code == 1009 && resp.Message == "cannot find compose" {
		return true
	}
	return false
}
