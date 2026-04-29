package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &licensingDataSource{}

func NewLicensingDataSource() datasource.DataSource {
	return &licensingDataSource{}
}

type licensingDataSource struct {
	client *client.Client
}

type licensingDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	ComputersCount         types.Int64  `tfsdk:"computers_count"`
	IOSCount               types.Int64  `tfsdk:"ios_count"`
	IPadOSCount            types.Int64  `tfsdk:"ipados_count"`
	MacOSCount             types.Int64  `tfsdk:"macos_count"`
	TVOSCount              types.Int64  `tfsdk:"tvos_count"`
	MaxDevices             types.Int64  `tfsdk:"max_devices"`
	TenantOverLicenseLimit types.Bool   `tfsdk:"tenant_over_license_limit"`
}

func (d *licensingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_licensing"
}

func (d *licensingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get detailed Iru tenant licensing information, including the total number of licensed slots and the current device utilization across platforms.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"computers_count":           schema.Int64Attribute{Computed: true},
			"ios_count":                 schema.Int64Attribute{Computed: true},
			"ipados_count":              schema.Int64Attribute{Computed: true},
			"macos_count":               schema.Int64Attribute{Computed: true},
			"tvos_count":                schema.Int64Attribute{Computed: true},
			"max_devices":               schema.Int64Attribute{Computed: true},
			"tenant_over_license_limit": schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *licensingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *licensingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data licensingDataSourceModel

	var licenseResp client.Licensing
	err := d.client.DoRequest(ctx, "GET", "/api/v1/settings/licensing", nil, &licenseResp)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read licensing, got error: %s", err))
		return
	}

	data.ComputersCount = types.Int64Value(int64(licenseResp.Counts.ComputersCount))
	data.IOSCount = types.Int64Value(int64(licenseResp.Counts.IOSCount))
	data.IPadOSCount = types.Int64Value(int64(licenseResp.Counts.IPadOSCount))
	data.MacOSCount = types.Int64Value(int64(licenseResp.Counts.MacOSCount))
	data.TVOSCount = types.Int64Value(int64(licenseResp.Counts.TVOSCount))
	data.MaxDevices = types.Int64Value(int64(licenseResp.Limits.MaxDevices))
	data.TenantOverLicenseLimit = types.BoolValue(licenseResp.TenantOverLicenseLimit)
	data.ID = types.StringValue("licensing")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
