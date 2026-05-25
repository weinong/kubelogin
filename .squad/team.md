# Squad Team

> kubelogin — `kubectl` credential plugin for non-interactive Azure AD / Entra ID authentication to Kubernetes clusters.

## Coordinator

| Name | Role | Notes |
|------|------|-------|
| Squad | Coordinator | Routes work, enforces handoffs and reviewer gates. |

## Members

| Name | Role | Charter | Status |
|------|------|---------|--------|
| 🏗️ Morpheus | Architect | `.squad/agents/morpheus/charter.md` | active |
| 🔧 Neo | Coder | `.squad/agents/neo/charter.md` | active |
| 🔍 Trinity | Reviewer | `.squad/agents/trinity/charter.md` | active |
| 📋 Scribe | Session Logger | `.squad/agents/scribe/charter.md` | active |
| 🔄 Ralph | Work Monitor | `.squad/agents/ralph/charter.md` | active |

## Project Context

- **Owner:** Weinong Wang
- **Project:** kubelogin
- **Description:** A Kubernetes `kubectl` credential (exec) plugin enabling non-interactive Azure AD / Entra ID authentication. Supports multiple auth modes: devicecode, interactive, spn, workload identity, msi, ropc, azurecli.
- **Stack:** Go, Kubernetes client-go exec credential plugin protocol, MSAL-Go, Azure AD / Entra ID OIDC
- **Created:** 2026-05-25
