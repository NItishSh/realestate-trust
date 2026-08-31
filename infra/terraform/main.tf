resource "kind_cluster" "default" {
  name           = var.cluster_name
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"

      kubeadm_config_patches = [
        <<-EOT
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
        EOT
      ]

      extra_port_mappings {
        container_port = 30000
        host_port      = 3000
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30080
        host_port      = 8080
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30081
        host_port      = 8081
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30082
        host_port      = 8082
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30083
        host_port      = 8083
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30084
        host_port      = 8084
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 30085
        host_port      = 8085
        protocol       = "TCP"
      }
    }

    node {
      role = "worker"
    }
  }
}

resource "null_resource" "build_and_load_images" {
  depends_on = [kind_cluster.default]

  triggers = {
    cluster_id = kind_cluster.default.id
  }

  provisioner "local-exec" {
    command = "bash ${path.module}/scripts/build_images.sh ${var.cluster_name} ${abspath(path.module)}/${var.project_root}"
  }
}

resource "kubernetes_namespace" "realestate_trust" {
  depends_on = [kind_cluster.default]
  metadata {
    name = "realestate-trust"
  }
}

resource "kubernetes_namespace" "argocd" {
  depends_on = [kind_cluster.default]
  metadata {
    name = "argocd"
  }
}

resource "helm_release" "argocd" {
  depends_on = [kubernetes_namespace.argocd, null_resource.build_and_load_images]

  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = "6.7.1"
  namespace        = "argocd"
  create_namespace = false

  set {
    name  = "server.extraArgs"
    value = "{--insecure}"
  }
}

resource "null_resource" "argocd_root_app" {
  depends_on = [helm_release.argocd]

  triggers = {
    cluster_id = kind_cluster.default.id
  }

  provisioner "local-exec" {
    command = "KUBECONFIG=${kind_cluster.default.kubeconfig_path} kubectl apply -f ${abspath(path.module)}/../gitops/root-application.yaml"
  }
}

resource "null_resource" "vault_eso_deployment" {
  depends_on = [null_resource.argocd_root_app]

  triggers = {
    cluster_id = kind_cluster.default.id
  }

  provisioner "local-exec" {
    # Using kubeconfig from cluster
    command = "KUBECONFIG=${kind_cluster.default.kubeconfig_path} bash ${abspath(path.module)}/../kind/deploy-vault-eso.sh"
  }
}
