package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/mohamed-mfarag/qnap-client-lib"
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
	Status            basetypes.StringValue `tfsdk:"status"`
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
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9_-]{0,30}[a-zA-Z0-9])?$`), "Application name must be between 2 and 32 characters, Valid characters: letters (a-z), numbers (0-9), hyphen (-), underscore (_)"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Required:    true,
				Description: "The state of the application (running, stopped). important to note that change in status requires complete recreation of the application - will be updated in the next version.",
				Validators: []validator.String{
					stringvalidator.OneOf("running", "stopped"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"yml": schema.StringAttribute{
				Required:    true,
				Description: "The YAML configuration for the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"removeanonvolumes": schema.BoolAttribute{
				Required:    true,
				Description: "Whether to remove anonymous volumes when the application is removed.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"containers": schema.ListNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the container.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the container.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				Description: "The list of containers in the application.",
			},
			"default_url": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The default URL for the application.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"port": schema.Int32Attribute{
						Optional:    true,
						Description: "The port number for the default URL.",
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.UseStateForUnknown(),
							int32planmodifier.RequiresReplace(),
						},
					},
					"service": schema.StringAttribute{
						Optional:    true,
						Description: "The service name for the default URL.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"cpu_limit": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The CPU limit for the application.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
					int32planmodifier.RequiresReplace(),
				},
			},
			"mem_limit": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory limit for the application.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
					int32planmodifier.RequiresReplace(),
				},
			},
			"mem_reservation": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The memory reservation for the application.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
					int32planmodifier.RequiresReplace(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "The last updated timestamp of the application.",
			},
		},
	}
}

// Create a new resource.
func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan, state *AppSpecModel

	newAppPlan, diags := ReadState(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new app
	app, err := r.client.CreateApplication(newAppPlan, &r.client.Token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	state, diags = GetCurrentState(ctx, plan, app)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// special handling for the removeanonvolumes attribute
	state.RemoveAnonVolumes = plan.RemoveAnonVolumes

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var priorState, newState *AppSpecModel
	diags := req.State.Get(ctx, &priorState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed application value from QNAP
	currentState, err := r.client.InspectApplication(priorState.Name.ValueString(), &r.client.Token)
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

	// Check if state is matching or not and return new status
	newState, diags = GetCurrentState(ctx, priorState, currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the RemoveAnonVolumes is equal
	newState.RemoveAnonVolumes = priorState.RemoveAnonVolumes
	// Set refreshed state

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete removes the resource from the Terraform state.
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

// Helper function to validate the YAML and convert it to String
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

// Helper function to check if the error is due to the application not being found
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

// Helper functions to read the state and convert it to the required format
func ReadState(ctx context.Context, state tfsdk.Plan) (qnap.NewAppReqModel, diag.Diagnostics) {
	// Retrieve values from plan
	var plan *AppSpecModel
	var diagnostics diag.Diagnostics
	diags := state.Get(ctx, &plan)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return qnap.NewAppReqModel{}, diagnostics
	}

	// Validate and convert YAML to JSON
	jsonString, err := validateYAML(plan.Yml.ValueString())
	if err != nil {
		diagnostics.AddError("error validating and converting YAML to string", err.Error())
		return qnap.NewAppReqModel{}, diagnostics
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
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return qnap.NewAppReqModel{}, diagnostics
		}
		// Handle default url attributes
		if default_url.Service.IsUnknown() || default_url.Service.IsNull() ||
			default_url.Port.IsUnknown() || default_url.Port.IsNull() {
			diagnostics.AddWarning("missing default url attributes", "Default URL is present however some attributes are missing, ensure service and port are present within the default url attribute.")
		} else {
			newApp.DefaultURL = qnap.NewAppReqDefaultURLModel{
				Service: default_url.Service.ValueString(),
				Port:    default_url.Port.ValueInt32(),
			}
		}
	}
	return newApp, diagnostics
}

// Helper function to compare the old state with the current state and generate a final state
func GetCurrentState(ctx context.Context, priorState *AppSpecModel, currentState *qnap.AppRespModel) (*AppSpecModel, diag.Diagnostics) {

	// Map response attributes to priorState attributes
	var newState *AppSpecModel = &AppSpecModel{}
	var priorStateCompose, currentStateCompose *ComposeFile
	var diagnostics diag.Diagnostics

	err := yaml.Unmarshal([]byte(priorState.Yml.ValueString()), &priorStateCompose)
	if err != nil {
		diagnostics.AddError("invalid YAML from priorState", err.Error())
		return nil, diagnostics
	}
	err = yaml.Unmarshal([]byte(currentState.Data.Yml), &currentStateCompose)
	if err != nil {
		diagnostics.AddError("invalid YAML from QNAP", err.Error())
		return nil, diagnostics
	}
	// Check if the compose files are equal - Usually does not change.
	if cmp.Equal(&currentStateCompose, &priorStateCompose) {
		newState.Yml = priorState.Yml
	} else {
		newState.Yml = types.StringValue(currentState.Data.Yml)
	}
	// Check if the CPU limit is equal
	if priorState.CPULimit.Equal(types.Int32Value(currentState.Data.CPULimit)) {
		newState.CPULimit = priorState.CPULimit
	} else {
		newState.CPULimit = types.Int32Value(currentState.Data.CPULimit)
	}
	// Check if the Mem limit is equal
	if priorState.MemLimit.Equal(types.Int32Value(currentState.Data.MemLimit)) {
		newState.MemLimit = priorState.MemLimit
	} else {
		newState.MemLimit = types.Int32Value(currentState.Data.MemLimit)
	}
	// Check if the Mem reservation is equal
	if priorState.MemReservation.Equal(types.Int32Value(currentState.Data.MemReservation)) {
		newState.MemReservation = priorState.MemReservation
	} else {
		newState.MemReservation = types.Int32Value(currentState.Data.MemReservation)
	}
	// Check if the status is equal
	if priorState.Status.Equal(types.StringValue(currentState.Data.Status)) {
		newState.Status = priorState.Status
	} else {
		newState.Status = types.StringValue(currentState.Data.Status)
	}

	// Convert []containers to basetypes.ListValue
	var containerListElements []attr.Value
	// Define the types for each attribute in the map
	containerAttrTypes := map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}
	for _, container := range currentState.Data.Containers {
		// Map the attributes' values
		containerMap := map[string]attr.Value{
			"name": types.StringValue(container.Name),
			"id":   types.StringValue(container.ID),
		}

		containerObject, diags := types.ObjectValue(containerAttrTypes, containerMap)
		diags.Append(diags...)
		if diags.HasError() {
			return nil, diags
		}

		containerListElements = append(containerListElements, containerObject)
	}
	// build container priorState
	newState.Containers = basetypes.NewListValueMust(types.ObjectType{AttrTypes: containerAttrTypes}, containerListElements)

	if !priorState.DefaultURL.IsNull() && !priorState.DefaultURL.IsUnknown() {
		var defaultURL DefaultURLModel
		diags := priorState.DefaultURL.As(ctx, &defaultURL, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		diags.Append(diags...)
		if diags.HasError() {
			return nil, diags
		}

		var newDefaultURL DefaultURLModel
		diags = newState.DefaultURL.As(ctx, &newDefaultURL, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		diags.Append(diags...)
		if diags.HasError() {
			return nil, diags
		}
		if defaultURL.Service.IsUnknown() || defaultURL.Service.IsNull() {
			newDefaultURL.Service = defaultURL.Service
		} else {
			newDefaultURL.Service = types.StringValue(currentState.Data.DefaultURL.Service)
		}
		if defaultURL.Port.IsUnknown() || defaultURL.Port.IsNull() {
			newDefaultURL.Port = defaultURL.Port
		} else {
			newDefaultURL.Port = types.Int32Value(currentState.Data.DefaultURL.Port)
		}
		// Convert newDefaultURL to basetypes.MapValue
		newDefaultURLAttrTypes := map[string]attr.Type{
			"service": types.StringType,
			"port":    types.Int32Type,
		}
		newDefaultURLMap := map[string]attr.Value{
			"service": types.StringValue(currentState.Data.DefaultURL.Service),
			"port":    types.Int32Value(currentState.Data.DefaultURL.Port),
		}

		defaultURLObject, diags := types.ObjectValue(newDefaultURLAttrTypes, newDefaultURLMap)
		diags.Append(diags...)
		if diags.HasError() {
			return nil, diags
		}
		newState.DefaultURL = defaultURLObject
	} else {
		newState.DefaultURL = priorState.DefaultURL
	}
	// Name must be equal as it stand as ID
	newState.Name = priorState.Name
	newState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return newState, diagnostics
}

// func CompareStates(ctx context.Context, priorState, currentState *AppSpecModel) (*AppSpecModel, diag.Diagnostics) {

// 	var operation basetypes.StringValue
// 	diagnostics := diag.Diagnostics{}
// 	// Retrieve values from priorState
// 	// var priorState AppSpecModel
// 	// diags := req.Plan.Get(ctx, &priorState)
// 	// diagnostics.Append(diags...)
// 	// if diagnostics.HasError() {
// 	// 	return nil, diagnostics
// 	// }

// 	// // Retrieve values from currentState
// 	// var currentState AppSpecModel
// 	// diags = req.State.Get(ctx, &currentState)
// 	// diagnostics.Append(diags...)
// 	// if diagnostics.HasError() {
// 	// 	return nil, diagnostics
// 	// }

// 	// Retrieve DefaultURL from currentState
// 	var currentStateDefaultURL DefaultURLModel
// 	diags := currentState.DefaultURL.As(ctx, &currentStateDefaultURL, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
// 	diagnostics.Append(diags...)
// 	if diagnostics.HasError() {
// 		return nil, diagnostics
// 	}

// 	// Retrieve DefaultURL from priorState
// 	var priorStateDefaultURL DefaultURLModel
// 	diags = priorState.DefaultURL.As(ctx, &priorStateDefaultURL, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
// 	diagnostics.Append(diags...)
// 	if diagnostics.HasError() {
// 		return nil, diagnostics
// 	}

// 	if priorState.Name != currentState.Name ||
// 		priorState.Yml != currentState.Yml ||
// 		priorState.RemoveAnonVolumes != currentState.RemoveAnonVolumes ||
// 		priorState.Status != currentState.Status ||
// 		priorState.CPULimit != currentState.CPULimit ||
// 		priorState.MemLimit != currentState.MemLimit ||
// 		priorState.MemReservation != currentState.MemReservation ||
// 		priorStateDefaultURL.Port != currentStateDefaultURL.Port ||
// 		priorStateDefaultURL.Service != currentStateDefaultURL.Service {
// 		//The difference between the create and update
// 		operation = types.StringValue("recreate")
// 	} else {
// 		priorState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
// 		// Set currentState to fully populated data
// 		diags = resp.State.Set(ctx, priorState)
// 		diagnostics.Append(diags...)
// 		if diagnostics.HasError() {
// 			return nil, diagnostics
// 		}
// 	}

// 	// Validate and convert YAML to JSON
// 	jsonString, err := validateYAML(priorState.Yml.ValueString())
// 	if err != nil {
// 		diagnostics.AddError(
// 			"Error validating and converting YAML to JSON",
// 			err.Error(),
// 		)
// 		return nil, diagnostics
// 	}

// 	newApp := qnap.NewAppReqModel{
// 		Name:           priorState.Name.ValueString(),
// 		Yml:            jsonString,
// 		CPULimit:       priorState.CPULimit.ValueInt32(),
// 		MemLimit:       priorState.MemLimit.ValueInt32(),
// 		MemReservation: priorState.MemReservation.ValueInt32(),
// 		Operation:      operation.ValueString(),
// 	}

// 	if !priorState.DefaultURL.IsNull() && !priorState.DefaultURL.IsUnknown() {
// 		var default_url DefaultURLModel
// 		diags := priorState.DefaultURL.As(ctx, &default_url, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
// 		diagnostics.Append(diags...)
// 		if diagnostics.HasError() {
// 			return nil, diagnostics
// 		}

// 		if default_url.Service.IsUnknown() || default_url.Service.IsNull() ||
// 			default_url.Port.IsUnknown() || default_url.Port.IsNull() {
// 			diagnostics.AddWarning("missing default url attributes", "Default URL is present however some attributes are missing, ensure service and port are present within the default url attribute.")
// 		} else {
// 			newApp.DefaultURL = qnap.NewAppReqDefaultURLModel{
// 				Service: default_url.Service.ValueString(),
// 				Port:    default_url.Port.ValueInt32(),
// 			}
// 		}
// 	}

// 	// Create new app
// 	app, err := r.client.CreateApplication(newApp, &r.client.Token)
// 	if err != nil {
// 		diagnostics.AddError(
// 			"Error creating app",
// 			"Could not create app, unexpected error: "+err.Error(),
// 		)
// 		return nil, diagnostics
// 	}

// 	// Map response body to schema and populate Computed attribute values

// 	priorState.CPULimit = types.Int32Value(app.Data.CPULimit)
// 	priorState.MemReservation = types.Int32Value(app.Data.MemReservation)
// 	priorState.MemLimit = types.Int32Value(app.Data.MemLimit)

// 	// Convert []containers to basetypes.ListValue
// 	var containerListElements []attr.Value
// 	// Define the types for each attribute in the map
// 	containerAttrTypes := map[string]attr.Type{
// 		"name": types.StringType,
// 		"id":   types.StringType,
// 	}
// 	for _, container := range app.Data.Containers {
// 		// Map the attributes' values
// 		containerMap := map[string]attr.Value{
// 			"name": types.StringValue(container.Name),
// 			"id":   types.StringValue(container.ID),
// 		}

// 		containerObject, diags := types.ObjectValue(containerAttrTypes, containerMap)
// 		diagnostics.Append(diags...)
// 		if diagnostics.HasError() {
// 			return nil, diagnostics
// 		}

// 		containerListElements = append(containerListElements, containerObject)
// 	}
// 	priorState.Containers = basetypes.NewListValueMust(types.ObjectType{AttrTypes: containerAttrTypes}, containerListElements)
// 	//validate the currentState matches what is expected in the priorState
// 	if priorState.Status.ValueString() == "running" {
// 		_, err = r.client.StartApplication(priorState.Name.ValueString(), &r.client.Token)
// 		if err != nil {
// 			diagnostics.AddError(
// 				"Error change application currentState to match requested currentState",
// 				"Could not start application, unexpected error: "+err.Error(),
// 			)
// 			return nil, diagnostics
// 		}
// 	} else if priorState.Status.ValueString() == "stopped" {
// 		_, err = r.client.StopApplication(priorState.Name.ValueString(), &r.client.Token)
// 		if err != nil {
// 			diagnostics.AddError(
// 				"Error change application currentState to match requested currentState",
// 				"Could not stop application, unexpected error: "+err.Error(),
// 			)
// 			return nil, diagnostics
// 		}
// 	}
// 	priorState.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
// 	// Set currentState to fully populated data
// 	diags = resp.State.Set(ctx, priorState)
// 	diagnostics.Append(diags...)
// 	if diagnostics.HasError() {
// 		return nil, diagnostics
// 	}
// }
