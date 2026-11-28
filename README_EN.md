# 🦫 Beaver IM - Enterprise-Grade Instant Messaging Platform

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-3.0.4-blue.svg)](VERSION)
[![Go](https://img.shields.io/badge/go-1.26.8-blue.svg)](https://golang.org/)
[![Go-Zero](https://img.shields.io/badge/Go--Zero-v1.7.4-green.svg)](https://github.com/zeromicro/go-zero)
[![QQ Group](https://img.shields.io/badge/QQ-1013328597%2B-blue.svg)](https://qm.qq.com/q/82rbf7QBzO)

> **Beaver IM core server** - Go-Zero microservices backing the Flutter mobile app and Electron desktop app with a complete, self-contained core IM closed loop (REST / WebSocket / gRPC). `beaver-server-package/` and `beaver-desktop-package/` are **independent, optional value-added services / capability packs**—core IM works normally even without them.

**Current version: [3.0.4](VERSION)** (see [`VERSION`](VERSION) at the repository root).

[English](README_EN.md) | [中文](README.md)

> For development and AI collaboration, read the [Chinese engineering handbook](docs/README.md) first. `.api` / `.proto` files are the source of truth for contracts; goctl generates the wiring. For older docs and examples, the current source code and handbook take precedence.

---

## 🌟 Key Features

- 🔐 **Authentication** - phone / email / verification code / QR code / OAuth login; the gateway verifies external JWTs and signs an ES256 assertion, services verify it with a public JWKS; S2S machine tickets, device management and remote logout
- 💬 **Messaging** - one-to-one and group chats, conversations and message history, read receipts, forwarding, WebSocket delivery and RocketMQ async pipelines
- 👥 **Social** - friend requests and profiles, group creation and moderation, block list; circles and moments live in the separate `beaver-social` repo
- 📁 **Files & Emoji** - uploads/downloads, emoji packs and favorites, with MinIO / Qiniu object storage
- 📞 **Calls** - LiveKit-backed signaling and rooms (`beaver-call`)
- 🔔 **Notifications & Push** - system notifications, interaction reminders, inbox and read state; a standalone `push_api` gateway with realtime / offline channels mapped to MQ topics
- 🔄 **Sync** - `datasync` incremental multi-device sync for friends, groups, messages, emoji, notifications, circles and bots
- 🏢 **Organization** - departments, employees, workbench, feedback / reports, client telemetry and release management in `platform`
- 🤖 **AI & Knowledge** - digital partners / bots, knowledge sources and skills (`SKILL.md`), agent conversation sync (`beaver-server-ai` + open platform)
- 📄 **Online Documents** - HTTP API and Yjs WebSocket in one `office_api` process (`beaver-server-online`)
- 📊 **Operations** - client logs flow through RocketMQ into OpenSearch; OpenSearch authentication is enabled with Basic Auth support
- 🛠️ **Admin** - user / organization / release admin endpoints in `backend_admin` (`beaver-server-backend`)

## 🛠️ Technology Stack

| Area | Technology | Notes |
|------|------------|-------|
| Language / framework | Go 1.26.8, go-zero 1.7.4 | goctl is a separate tool; its version is not pinned in the repo (see handbook 1.1) |
| Communication | gRPC, HTTP S2S, WebSocket, RocketMQ | RPC inside this repo; HTTP machine tickets across repos and for push |
| Storage / search | MySQL 8.0, Redis 6.0, ETCD 3.5, OpenSearch | Split databases, cache, service discovery, log search |
| Deployment | Docker Compose, build scripts | `deploy/<domain>/` and `docker/` |
| Other | JWT (ES256 / JWKS), LiveKit, MinIO / Qiniu | Auth, calls, object storage |

## 🏗️ Architecture

Clients (Flutter / Electron) reach the backend through `gateway_api :20800`. The gateway verifies external Beaver / Open / Admin tickets, signs an ES256 assertion, injects trusted identity headers and forwards by path. Services verify the assertion with JWKS; cross-service and cross-repo calls use S2S machine tickets and `machine` / `_public` views. WebSocket carries long-lived connections; RocketMQ carries async messages and push delivery. The admin console reaches `backend_admin` through the gateway.

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Flutter app  │  │ Electron app │  │ Admin console│
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       └─────────────────┼─────────────────┘
                         │
              ┌──────────▼──────────┐
              │  gateway_api :20800  │
              └──────────▼──────────┘
                         │
   ┌─────────────────────┼─────────────────────┐
   │                     │                     │
account / relation   message / ws          asset / notification
platform / datasync      push :20500        (RPC inside this repo only)
   │                     │                     │
   └─────────────────────┼─────────────────────┘
                         │
        ┌────────────────┴────────────────┐
        │ MySQL · Redis · ETCD · OpenSearch │
        └─────────────────────────────────┘

  Cross-repo ──► social / call / open / office / agent / backend (HTTP machine views)
```

### Service Ports

[`server.md`](server.md) is the single port table; actual listeners are in each service's `etc/*.yaml`. The main repo has 9 business domains plus the standalone push service. Subrepos run single processes; old moment / circle / call_rpc / open_rpc / doc_rpc ports are no longer in use.

| Service | API | RPC | Repo | Notes |
|---------|----:|----:|------|-------|
| gateway_api | 20800 | - | beaver-server | Client API gateway: ticket verification and forwarding |
| account | 20100 | 30100 | beaver-server | Users + auth (merged) |
| relation | 20200 | 30200 | beaver-server | Friends + groups (merged) |
| message | 20300 | 30300 | beaver-server | Chat (renamed from chat) |
| ws | 20400 | - | beaver-server | WebSocket connections |
| push | 20500 | - | beaver-server | Standalone push gateway (S2S) |
| asset | 20600 | 30600 | beaver-server | Files + emoji (merged) |
| notification | 21000 | 31000 | beaver-server | Notifications |
| platform | 21600 | 31100 | beaver-server | Org / feedback / telemetry / updates |
| datasync | 21700 | - | beaver-server | Incremental sync |
| call | 21800 | - | beaver-call | Calls (LiveKit, single process) |
| open | 21900 | - | beaver-open | Open platform: API / portal / WS / MQ |
| social | 22100 | - | beaver-social | Circles + moments (single process, two DBs) |
| agent | 22300 | - | beaver-server-ai | AI / digital partners (Python, `agent_api`, optional) |
| office | 22400 | - | beaver-server-online | Online docs (HTTP + Yjs WS, optional) |
| observability | 22800 | - | beaver-server-observability | Logging & observability (OpenSearch + Vector, optional) |
| backend_admin | 40000 | - | beaver-server-backend | Admin APIs (JWT / CORS, optional) |

## 📁 Project Structure

```
beaver-server/
├── app/                    # 9 domains: account relation message asset
│   │                       # notification platform datasync ws gateway (+ push)
│   ├── account/            # account_api + account_rpc + account_models
│   ├── relation/           # relation_api + relation_rpc + relation_models
│   ├── message/            # message_api + message_rpc + message_models
│   ├── asset/              # asset_api + asset_rpc + asset_models
│   ├── notification/       # notification_api + notification_rpc + notification_models
│   ├── platform/           # platform_api + platform_rpc + platform_models
│   ├── datasync/           # datasync_api + datasync_models (no RPC)
│   ├── ws/                 # ws_api (no RPC)
│   ├── push/               # push_api (no RPC, no tables)
│   └── gateway/            # gateway_api (custom proxy, no .api)
├── common/                 # Shared constants, identity, response, middleware
├── core/                   # MySQL / Redis / MQ / etcd / OpenSearch / push clients
├── commonconfig/           # Shared config structs
├── middleware/             # JWT and HTTP middleware
├── remote/                 # Outbound clients for other repos, grouped by peer
├── types/                  # Outbound request / response DTOs
├── openevent/              # Public event publishing
├── utils/                  # Logging, JWT, S2S, password, email helpers
├── template/               # Repo-specific goctl templates
├── database/               # DB seed data (not SQL migrations)
├── deploy/                 # Compose and config per domain
├── build/                  # Build / release / local start scripts
├── docs/                   # Engineering handbook (Chinese, 13 chapters)
├── main.go                 # DB init tool only: go run main.go -db
└── server.md               # Service port reference
```

## 🚀 Quick Start

See the [backend docs](https://wsrh8888.github.io/beaver-docs/backend/) and [deployment guide](https://wsrh8888.github.io/beaver-docs/backend/deploy/build-scripts.html), plus the local handbook chapter [7 Configuration & Deployment](docs/07-deploy/README.md).

Initialize the local database (MySQL connection must be configured first; `main.go` is **a DB init tool, not a service entry point**):

```bash
go run main.go -db
```

Build and orchestration scripts live in `build/` and `deploy/<domain>/`. `VERSION` is used by the build scripts for image tags.

## 📚 Documentation & Resources

- 📖 [Engineering handbook](docs/README.md) (Chinese: architecture, contracts, boundaries, identity, data, deployment, decisions)
- 📖 [Full documentation](https://wsrh8888.github.io/beaver-docs/)
- 📝 [Changelog](CHANGELOG.md) · [Update notes](update.md)
- 🎥 [Video tutorials](https://space.bilibili.com/269553626/lists)
- 📱 [Android demo APK](https://github.com/wsrh8888/beaver-docs/releases/download/lastest/latest.apk)
- 💬 [QQ Group: 1013328597](https://qm.qq.com/q/82rbf7QBzO)

## 🔗 Related Projects

All repositories live under the GitHub `beaver-im` organization. **The core repository forms a complete, self-contained IM closed loop on its own**; `beaver-server-package/` and `beaver-desktop-package/` are **independent, optional value-added services / capability packs** (local container folders, not repositories themselves)—core IM works normally even without them.

| Group | Project | Repository | Description |
|-------|---------|------------|-------------|
| **Core Server** | beaver-server | [GitHub](https://github.com/beaver-im/beaver-server) | Core backend monorepo (self-contained closed loop, this repo) |
| **Core Clients** | beaver-flutter | [GitHub](https://github.com/beaver-im/beaver-flutter) | Mobile app (Flutter, recommended) |
| | beaver-desktop | [GitHub](https://github.com/beaver-im/beaver-desktop) | Desktop app core (Electron) |
| **Portals / SDK / Docs** | beaver-manager | [GitHub](https://github.com/beaver-im/beaver-manager) | Admin web console |
| | beaver-oauth | [GitHub](https://github.com/beaver-im/beaver-oauth) | OAuth login portal |
| | beaver-open | [GitHub](https://github.com/beaver-im/beaver-open) | Open platform web portal |
| | beaver-sdk | [GitHub](https://github.com/beaver-im/beaver-sdk) | JS SDK & OpenClaw adapter |
| | beaver-docs | [GitHub](https://github.com/beaver-im/beaver-docs) | Official documentation site |
| **Optional Server Add-ons** | beaver-server-ai | [GitHub](https://github.com/beaver-im/beaver-server-ai) | AI agent (Python, optional) |
| (`beaver-server-package`) | beaver-server-backend | [GitHub](https://github.com/beaver-im/beaver-server-backend) | Admin APIs (`backend_admin`, optional) |
| | beaver-server-call | [GitHub](https://github.com/beaver-im/beaver-server-call) | Calls (LiveKit, optional) |
| | beaver-server-observability | [GitHub](https://github.com/beaver-im/beaver-server-observability) | Logging & observability (optional) |
| | beaver-server-online | [GitHub](https://github.com/beaver-im/beaver-server-online) | Online docs (`office_api`, optional) |
| | beaver-server-open | [GitHub](https://github.com/beaver-im/beaver-server-open) | Open platform / bots (optional) |
| | beaver-server-social | [GitHub](https://github.com/beaver-im/beaver-server-social) | Circles and moments (`social_api`, optional) |
| **Optional Desktop Packs** | beaver-desktop-ai | [GitHub](https://github.com/beaver-im/beaver-desktop-ai) | External desktop AI capability pack (optional) |
| (`beaver-desktop-package`) | beaver-desktop-online | [GitHub](https://github.com/beaver-im/beaver-desktop-online) | External desktop online-docs pack (optional) |

## 📱 Feature Showcase

### 🔐 Authentication & Security
<div align="center">
  <img src="./static/mobile/login.jpg" width="200" alt="Login Interface"/>
  <img src="./static/mobile/register.jpg" width="200" alt="Registration Interface"/>
  <img src="./static/mobile/find-password.jpg" width="200" alt="Password Recovery"/>
</div>

### 💬 Messaging & Communication
<div align="center">
  <img src="./static/mobile/message.jpg" width="200" alt="Message Interface"/>
  <img src="./static/mobile/private-chat.jpg" width="200" alt="Private Chat"/>
  <img src="./static/mobile/group-chat.jpg" width="200" alt="Group Chat"/>
  <img src="./static/mobile/send-text.jpg" width="200" alt="Send Text"/>
  <img src="./static/mobile/send-emoji.jpg" width="200" alt="Send Emoji"/>
  <img src="./static/mobile/chat-details.jpg" width="200" alt="Chat Details"/>
</div>

### 👥 Social Features
<div align="center">
  <img src="./static/mobile/friend.jpg" width="200" alt="Friend List"/>
  <img src="./static/mobile/new-friends.jpg" width="200" alt="New Friends"/>
  <img src="./static/mobile/friend-info.jpg" width="200" alt="Friend Profile"/>
  <img src="./static/mobile/edit-remark.jpg" width="200" alt="Edit Remark"/>
</div>

### 🏠 Moments & Groups
<div align="center">
  <img src="./static/mobile/moments.jpg" width="200" alt="Moments"/>
  <img src="./static/mobile/send-moments.jpg" width="200" alt="Send Moments"/>
  <img src="./static/mobile/group-list.jpg" width="200" alt="Group List"/>
  <img src="./static/mobile/create-group.jpg" width="200" alt="Create Group"/>
  <img src="./static/mobile/group-details.jpg" width="200" alt="Group Details"/>
  <img src="./static/mobile/add-members.jpg" width="200" alt="Add Members"/>
</div>

### 👤 User Management
<div align="center">
  <img src="./static/mobile/mine.jpg" width="200" alt="User Center"/>
  <img src="./static/mobile/profile-edit.jpg" width="200" alt="Profile Editing"/>
  <img src="./static/mobile/qcode.jpg" width="200" alt="QR Code Features"/>
</div>

### ⚙️ System Features
<div align="center">
  <img src="./static/mobile/settings.jpg" width="200" alt="Settings"/>
  <img src="./static/mobile/update.jpg" width="200" alt="Update"/>
  <img src="./static/mobile/feedback.jpg" width="200" alt="Feedback"/>
  <img src="./static/mobile/about.jpg" width="200" alt="About"/>
  <img src="./static/mobile/statement.jpg" width="200" alt="Statement"/>
</div>

## 🤝 Contributing

Issues and pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License & Legal Disclaimer

This project is licensed under the [MIT License](LICENSE) — see [LICENSE](LICENSE).

**Summary:**

- Closed-source self-use and re-open-sourcing derivatives are free if you keep the root `LICENSE` and show About attribution when deployed (Based on Beaver IM + repo link)
- Closed-source delivery to third parties, removing About, or running a paid IM SaaS requires a commercial license (written contract)
- **Do not delete or alter `LICENSE`**, whether paid or free

Full disclaimer & attribution: [LEGAL.md](LEGAL.md)  
Commercial product lines & pricing: [Copyright & Commercial Licensing](https://wsrh8888.github.io/beaver-docs/community/license.html)  
Contact: [751135385@qq.com](mailto:751135385@qq.com)

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=beaver-im/beaver-server&type=Date)](https://star-history.com/#beaver-im/beaver-server&Date)

## ☕ Buy the Author a Coffee

<div align="center">
  <img src="./static/sponsor/wechat.jpg" width="200" alt="WeChat Sponsor QR Code"/>
  <img src="./static/sponsor/zhifubao.jpg" width="200" alt="Alipay Sponsor QR Code"/>
</div>

---

<div align="center">
  <strong>Made with ❤️ by Beaver IM Team</strong><br>
  <em>Enterprise-Grade Instant Messaging Platform</em>
</div>
