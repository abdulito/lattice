load(":bazel/repositories.bzl", "rules_go_dependencies", "rules_docker_dependencies", "rules_package_manager_dependencies")
rules_go_dependencies()
rules_docker_dependencies()
rules_package_manager_dependencies()

load(":bazel/initialize.bzl", "initialize_rules_go", "initialize_rules_docker", "initialize_rules_package_manager")
initialize_rules_go()
initialize_rules_docker()
initialize_rules_package_manager()

load(":bazel/dependencies.bzl", "go_dependencies", "docker_dependencies")
go_dependencies()
docker_dependencies()
