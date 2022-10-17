data "cloudfoundry_service" "service_test" {
    name = "xsuaa"
    # space = "28d38f56-c191-4923-8107-ab3d59e4ff53"
}

output "service_test_output" {
    value = data.cloudfoundry_service.service_test
}