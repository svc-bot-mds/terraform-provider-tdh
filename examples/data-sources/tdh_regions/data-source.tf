# pass valid data with respect to the instance type selected
data "tdh_regions" "dedicated_aws" {
  instance_size        = "XX-SMALL"
  dedicated_data_plane = true
}