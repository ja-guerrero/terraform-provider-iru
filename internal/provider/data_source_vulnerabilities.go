package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

var _ datasource.DataSource = &vulnerabilitiesDataSource{}

func NewVulnerabilitiesDataSource() datasource.DataSource {
	return &vulnerabilitiesDataSource{}
}

type vulnerabilitiesDataSource struct {
	client *client.Client
}

type vulnerabilitiesDataSourceModel struct {
	ID      types.String         `tfsdk:"id"`
	Results []vulnerabilityModel `tfsdk:"results"`
}

type vulnerabilityModel struct {
	CVEID              types.String  `tfsdk:"cve_id"`
	Severity           types.String  `tfsdk:"severity"`
	CVSSScore          types.Float64 `tfsdk:"cvss_score"`
	FirstDetectionDate types.String  `tfsdk:"first_detection_date"`
	DeviceCount        types.Int64   `tfsdk:"device_count"`
	Status             types.String  `tfsdk:"status"`
	Software           types.List    `tfsdk:"software"`
}

func (d *vulnerabilitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vulnerabilities"
}

func (d *vulnerabilitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all vulnerabilities from Vulnerability Management.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"results": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cve_id": schema.StringAttribute{
							Computed: true,
						},
						"severity": schema.StringAttribute{
							Computed: true,
						},
						"cvss_score": schema.Float64Attribute{
							Computed: true,
						},
						"first_detection_date": schema.StringAttribute{
							Computed: true,
						},
						"device_count": schema.Int64Attribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"software": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *vulnerabilitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *vulnerabilitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vulnerabilitiesDataSourceModel

	var all []client.Vulnerability
	page := 1
	size := 50

	for {
		type vulnerabilityResponse struct {
			Results []client.Vulnerability `json:"results"`
			Next    string                 `json:"next"`
		}
		var listResp vulnerabilityResponse

		path := fmt.Sprintf("/api/v1/vulnerability-management/vulnerabilities?size=%d&page=%d", size, page)
		err := d.client.DoRequest(ctx, "GET", path, nil, &listResp)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read vulnerabilities, got error: %s", err))
			return
		}

		all = append(all, listResp.Results...)

		if listResp.Next == "" || len(listResp.Results) < size {
			break
		}
		page++
	}

	for _, item := range all {
		softwareList, diags := types.ListValueFrom(ctx, types.StringType, item.Software)
		resp.Diagnostics.Append(diags...)

		data.Results = append(data.Results, vulnerabilityModel{
			CVEID:              types.StringValue(item.CVEID),
			Severity:           types.StringValue(item.Severity),
			CVSSScore:          types.Float64Value(item.CVSSScore),
			FirstDetectionDate: types.StringValue(item.FirstDetectionDate),
			DeviceCount:        types.Int64Value(int64(item.DeviceCount)),
			Status:             types.StringValue(item.Status),
			Software:           softwareList,
		})
	}

	data.ID = types.StringValue("vulnerabilities")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
