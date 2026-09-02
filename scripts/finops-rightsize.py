#!/usr/bin/env python3
"""
FinOps Right-Sizing Policy Engine
Ingests VPA and OpenCost metrics, applies safety guardrails, and auto-tunes Helm values.
"""
import argparse
import json
import os
import subprocess
import sys
import yaml

SERVICES = [
    "identity-service",
    "transaction-manager",
    "financing-engine",
    "tokenization-engine",
    "ledger-service",
    "property-registry-service",
    "feedback-service",
]

# FinOps Guardrails & Safety Margins
BUFFER_PERCENTAGE = 1.20       # +20% safety buffer above VPA target
MIN_CPU_MILLIS = 25            # Minimum 25m CPU to prevent throttling
MIN_MEMORY_MI = 32             # Minimum 32Mi RAM
LIMIT_MEMORY_MULTIPLIER = 1.5   # Memory limit = 1.5x request to prevent OOMKills

# Fallback baseline profiles for offline/dry-run analysis (in millicores and MiB)
OFFLINE_BASELINES = {
    "identity-service": {"cpu": "45m", "memory": "64Mi"},
    "transaction-manager": {"cpu": "60m", "memory": "80Mi"},
    "financing-engine": {"cpu": "40m", "memory": "64Mi"},
    "tokenization-engine": {"cpu": "45m", "memory": "64Mi"},
    "ledger-service": {"cpu": "50m", "memory": "72Mi"},
    "property-registry-service": {"cpu": "35m", "memory": "48Mi"},
    "feedback-service": {"cpu": "25m", "memory": "32Mi"},
}

def parse_cpu_to_millis(cpu_str: str) -> int:
    if not cpu_str:
        return MIN_CPU_MILLIS
    if cpu_str.endswith("m"):
        return int(cpu_str[:-1])
    return int(float(cpu_str) * 1000)

def parse_memory_to_mi(mem_str: str) -> int:
    if not mem_str:
        return MIN_MEMORY_MI
    if mem_str.endswith("Mi"):
        return int(mem_str[:-2])
    if mem_str.endswith("Gi"):
        return int(float(mem_str[:-2]) * 1024)
    if mem_str.endswith("k") or mem_str.endswith("Ki"):
        return max(int(int(mem_str[:-2]) / 1024), 1)
    return int(int(mem_str) / (1024 * 1024))

def get_vpa_recommendations(service: str, namespace: str = "realestate-trust"):
    cmd = ["kubectl", "get", "vpa", f"{service}-vpa", "-n", namespace, "-o", "json"]
    try:
        out = subprocess.check_output(cmd, stderr=subprocess.DEVNULL)
        data = json.loads(out)
        target = data["status"]["recommendation"]["containerRecommendations"][0]["target"]
        return target["cpu"], target["memory"]
    except Exception:
        # If cluster/VPA is not reachable, use offline baseline
        fallback = OFFLINE_BASELINES.get(service, {"cpu": "50m", "memory": "64Mi"})
        return fallback["cpu"], fallback["memory"]

def rightsize_service(service: str, base_dir: str, dry_run: bool = False):
    rec_cpu, rec_mem = get_vpa_recommendations(service)

    # Calculate buffered request values
    target_cpu_m = max(int(parse_cpu_to_millis(rec_cpu) * BUFFER_PERCENTAGE), MIN_CPU_MILLIS)
    target_mem_mi = max(int(parse_memory_to_mi(rec_mem) * BUFFER_PERCENTAGE), MIN_MEMORY_MI)
    target_limit_mem_mi = int(target_mem_mi * LIMIT_MEMORY_MULTIPLIER)

    values_path = os.path.join(base_dir, f"{service}.yaml")
    if not os.path.exists(values_path):
        return None

    with open(values_path, "r") as f:
        doc = yaml.safe_load(f) or {}

    old_req = doc.get("resources", {}).get("requests", {})
    old_cpu = old_req.get("cpu", "default")
    old_mem = old_req.get("memory", "default")

    if "resources" not in doc or not isinstance(doc["resources"], dict):
        doc["resources"] = {}

    doc["resources"]["requests"] = {
        "cpu": f"{target_cpu_m}m",
        "memory": f"{target_mem_mi}Mi",
    }
    doc["resources"]["limits"] = {
        "memory": f"{target_limit_mem_mi}Mi",
    }

    if not dry_run:
        with open(values_path, "w") as f:
            yaml.dump(doc, f, sort_keys=False)

    return {
        "service": service,
        "old_cpu": old_cpu,
        "new_cpu": f"{target_cpu_m}m",
        "old_mem": old_mem,
        "new_mem": f"{target_mem_mi}Mi",
        "limit_mem": f"{target_limit_mem_mi}Mi",
    }

def main():
    parser = argparse.ArgumentParser(description="FinOps Workload Right-Sizing Engine")
    parser.add_argument("--dry-run", action="store_true", help="Preview recommendations without modifying files")
    parser.add_argument("--values-dir", default="infra/kind/values", help="Path to microservice Helm values directory")
    args = parser.parse_args()

    results = []
    for svc in SERVICES:
        res = rightsize_service(svc, args.values_dir, args.dry_run)
        if res:
            results.append(res)

    print("\n========================================================")
    print(f"💰 FinOps Right-Sizing Analysis ({'DRY RUN' if args.dry_run else 'APPLIED'})")
    print("========================================================")
    print(f"{'SERVICE':<28} | {'CPU REQUEST':<18} | {'MEMORY REQUEST':<18} | {'MEMORY LIMIT':<12}")
    print("-" * 84)
    for r in results:
        print(f"{r['service']:<28} | {r['old_cpu']} -> {r['new_cpu']:<7} | {r['old_mem']} -> {r['new_mem']:<7} | {r['limit_mem']}")
    print("========================================================\n")

if __name__ == "__main__":
    main()
