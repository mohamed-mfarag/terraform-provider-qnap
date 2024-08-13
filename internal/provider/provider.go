package provider

import (
	"context"
	"os"

	"github.com/hashicorp/qnap-client-lib"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &qnapProvider{}
)

// qnapProviderModel maps provider schema data to a Go type.
type qnapProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &qnapProvider{
			version: version,
		}
	}
}

// qnapProvider is the provider implementation.
type qnapProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *qnapProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "qnap"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *qnapProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a qnap API client for data sources and resources.
func (p *qnapProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config qnapProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown qnap API Host",
			"The provider cannot create the qnap API client as there is an unknown configuration value for the qnap API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the qnap_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown qnap API Username",
			"The provider cannot create the qnap API client as there is an unknown configuration value for the qnap API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the qnap_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown qnap API Password",
			"The provider cannot create the qnap API client as there is an unknown configuration value for the qnap API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the qnap_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("QNAP_HOST")
	username := os.Getenv("QNAP_USERNAME")
	password := os.Getenv("QNAP_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing qnap API Host",
			"The provider cannot create the qnap API client as there is a missing or empty value for the qnap API host. "+
				"Set the host value in the configuration or use the QNAP_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing qnap API Username",
			"The provider cannot create the qnap API client as there is a missing or empty value for the qnap API username. "+
				"Set the username value in the configuration or use the QNAP_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing qnap API Password",
			"The provider cannot create the qnap API client as there is a missing or empty value for the qnap API password. "+
				"Set the password value in the configuration or use the QNAP_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new qnap client using the configuration values
	client, err := qnap.NewClient(&host, &username, &password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create qnap API Client",
			"An unexpected error occurred when creating the qnap API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"qnap Client Error: "+err.Error(),
		)
		return
	}

	// Make the qnap client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
// DataSources defines the data sources implemented in the provider.
func (p *qnapProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContainersDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *qnapProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewContainerResource,
		NewAppResource,
	}
}
