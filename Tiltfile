# -*- mode: Python -*-

# set defaults

kind_cluster_name = "hostport-allocator"
kind_kubeconfig = "kind-kubeconfig"

settings = {
    'preload_images_for_kind': True
}

# global settings
settings.update(read_json(
    "tilt-settings.json",
    default={},
))


def deploy_baremetal_manager():
    # Set up a local_resource build of the provider's manager binary. The provider is expected to have a main.go in
    # manager_build_path. The binary is written to .tiltbuild/manager.
    local_resource(
        "manager",
        cmd='make manager-amd64',
        deps=[
            "main.go",
            "api",
            "apis",
            "controllers",
            "webhook",
            "webhooks"
        ],
    )

    custom_build(
        'controller',
        'docker build -t $EXPECTED_REF .',
        deps=[
            'bin/hostport-allocator-manager-linux-amd64',
            'Dockerfile'
        ],
    )

    yaml = str(kustomize("config/tilt"))
    substitutions = settings.get("kustomize_substitutions", {})
    for substitution in substitutions:
        value = substitutions[substitution]
        yaml = yaml.replace("${" + substitution + "}", value)
    k8s_yaml(blob(yaml))


# Prepull all the cert-manager images to your local environment and then load them directly into kind. This speeds up
# setup if you're repeatedly destroying and recreating your kind cluster, as it doesn't have to pull the images over
# the network each time.
def deploy_cert_manager():
    registry = "quay.io/jetstack"
    # https://github.com/jetstack/cert-manager/issues/3121
    version = "v0.15.2"
    images = ["cert-manager-controller", "cert-manager-cainjector", "cert-manager-webhook"]

    if settings.get("preload_images_for_kind"):
        for image in images:
            local("docker pull {}/{}:{}".format(registry, image, version))
            local("kind load docker-image --name {} {}/{}:{}".format(kind_cluster_name, registry, image, version))

    local(
        "kubectl apply --kubeconfig {} -f https://github.com/jetstack/cert-manager/releases/download/{}/cert-manager.yaml".format(
            kind_kubeconfig, version))

    # wait for the service to become available
    local(
        "kubectl wait -n cert-manager --kubeconfig {} --for=condition=Available --timeout=300s deployment cert-manager-webhook".format(
            kind_kubeconfig))


# Users may define their own Tilt customizations in tilt.d. This directory is excluded from git and these files will
# not be checked in to version control.
def include_user_tilt_files():
    user_tiltfiles = listdir("tilt.d")
    for f in user_tiltfiles:
        include(f)


##############################
# Actual work happens here
##############################
include_user_tilt_files()

deploy_cert_manager()

deploy_baremetal_manager()
