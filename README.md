<div align="center">

<br/>

<img alt="vrcollab" src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=800&size=44&duration=2400&pause=900&color=A78BFA&center=true&vCenter=true&width=900&height=80&lines=vrcollab"/>

**Self-hosted multiplayer VR — your own metaverse server.**
_WebRTC SFU · Real-time pose sync · Spatial audio · Asset CDN · Unity + Unreal SDKs._

<br/>

<p>
<img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white"/>
<img src="https://img.shields.io/badge/WebRTC-333333?style=for-the-badge&logo=webrtc&logoColor=white"/>
<img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white"/>
<img src="https://img.shields.io/badge/Unity-FFFFFF?style=for-the-badge&logo=unity&logoColor=black"/>
<img src="https://img.shields.io/badge/Unreal%20Engine-313131?style=for-the-badge&logo=unrealengine&logoColor=white"/>
</p>

<p>
<img src="https://img.shields.io/github/stars/kasimmj/vrcollab?style=social"/>
<img src="https://img.shields.io/github/forks/kasimmj/vrcollab?style=social"/>
</p>

</div>

---

## 🥽 Why vrcollab?

Photon costs money. Mirror requires you to host. Unity Multiplayer keeps shutting down. Building your own VR multiplayer from scratch takes 3 months.

**vrcollab is the open-source, self-hosted answer:**
- 🌐 Up to 50 concurrent users per room (tested on a single 4-vCPU VPS)
- 🎙️ Spatial voice via WebRTC (LiveKit-compatible SFU)
- 🤸 Sub-60ms pose sync at 90Hz for hand + head tracking
- 📦 Built-in asset CDN with on-demand streaming
- 🏗️ Drop-in SDKs for Unity, Unreal Engine 5, and WebXR
- 🔒 End-to-end encryption optional
- 🛡️ Anti-cheat hooks built into the protocol

---

## ⚡ Quick Start

### 1. Run the server

```bash
git clone https://github.com/kasimmj/vrcollab
cd vrcollab
docker compose up -d
```

You now have:
- **Signaling/room manager** at `:7880`
- **SFU media servers** at `:50000-50100/udp`
- **Asset CDN** at `:8080`
- **Admin dashboard** at `:7881`

### 2. Drop the SDK into your project

**Unity:**
```bash
# Add via Unity Package Manager
git+https://github.com/kasimmj/vrcollab.git?path=/sdks/unity
```

```csharp
using VRCollab;

var room = await VRCollabClient.JoinRoom("metaverse-hilla", playerName: "Kasim");
room.OnUserJoined += user => Debug.Log($"{user.Name} joined");
room.OnUserPose += pose => UpdateAvatar(pose);
```

**Unreal Engine 5:**
Drop the `VRCollab` plugin into `Plugins/`, then:

```cpp
#include "VRCollabSubsystem.h"

auto* VRC = GetGameInstance()->GetSubsystem<UVRCollabSubsystem>();
VRC->OnUserJoined.AddDynamic(this, &AMyGameMode::HandleUserJoined);
VRC->JoinRoom("metaverse-hilla", "Kasim");
```

**WebXR:**
```js
import { VRCollab } from '@vrcollab/client';
const room = await VRCollab.join('metaverse-hilla', { name: 'Kasim' });
```

---

## 🏗️ Architecture

```
                ┌────────────────────────────────────────┐
                │           SIGNALING LAYER (Go)          │
                │  Room manager · auth · token issuance   │
                └─────────────┬──────────────────────────┘
                              │
                  ┌───────────┴───────────┐
                  │                       │
              ┌───▼──────┐         ┌──────▼────┐
              │  SFU #1  │ ... N   │  SFU #N   │
              │ (Pion)   │         │  (Pion)   │
              └────┬─────┘         └─────┬─────┘
                   │                     │
              ┌────▼─────────────────────▼─────┐
              │       VR Clients (Unity/UE)     │
              └─────────────────────────────────┘

         ┌──────────────┐   ┌──────────────┐
         │  Asset CDN   │   │   Postgres   │
         │  (S3/MinIO)  │   │  (rooms/users)│
         └──────────────┘   └──────────────┘
```

- **Signaling:** custom Go service over WebSocket — issues JWTs, manages room state
- **SFU:** Pion (Go) — relays audio + low-bandwidth pose streams between peers
- **Asset CDN:** MinIO + on-the-fly transcoding for textures and meshes
- **DB:** Postgres for room metadata, user accounts, and audit logs

---

## 🎙️ Spatial Audio

Audio is positioned in 3D space using **HRTF** (Head-Related Transfer Function) on the client. The server only routes streams — it doesn't process audio.

Configuration:
```yaml
audio:
  spatial: true
  hrtf_dataset: "ircam-2024"      # or "mit-kemar"
  max_distance: 30.0              # meters before silence
  rolloff: "inverse"              # "linear" | "inverse" | "exponential"
  voice_codec: "opus"             # 16kHz mono, 32kbps
```

---

## 🤸 Pose Sync Protocol

vrcollab uses a custom binary protocol over WebRTC data channels for pose sync. Every 11ms (90Hz), each client sends:

```
PoseFrame {
  user_id:    u32
  timestamp:  u64 (microseconds since session start)
  head:       Pose (position + rotation)
  left_hand:  Pose
  right_hand: Pose
  body_ik:    [16 × Pose]  (optional, for full-body)
  voice_amp:  f32          (for lip sync)
}
```

Total: ~120 bytes per frame, ~88 Kbps per user. The server uses **dead reckoning** to predict missed packets and **delta encoding** to reduce bandwidth by ~60%.

---

## 🛡️ Anti-Cheat Hooks

vrcollab provides protocol-level hooks:

- **Server-side teleport validation** — refuse poses with sub-second jumps > 5m
- **Voice amplitude clipping** — prevent voice volume exploits
- **Rate limiting** per RPC, per user
- **Replay protection** via sequence numbers + timestamps

Game-specific anti-cheat (aimbot detection, animation cancels) is up to your SDK integration.

---

## 📊 Capacity & Performance

Benchmarked on a 4-vCPU 8GB VPS (single node):

| Metric | Value |
|--------|-------|
| Max concurrent users in one room | 50 |
| Max simultaneous rooms | 200 |
| Total bandwidth at 50 users | ~5 Mbps |
| End-to-end pose latency (LAN) | 22ms |
| End-to-end pose latency (WAN US-EU) | 95ms |
| Audio glass-to-glass latency | 110ms |

For larger deployments, vrcollab supports horizontal SFU scaling — add more `sfu` containers.

---

## ⚙️ Configuration

`config.yaml`:

```yaml
server:
  signaling_port: 7880
  admin_port: 7881
  jwt_secret: "${JWT_SECRET}"

sfu:
  rtp_ports: 50000-50100
  ice_lite: false
  external_ip: "${PUBLIC_IP}"

rooms:
  default_capacity: 30
  max_capacity: 100
  idle_timeout: 5m

assets:
  cdn_url: "https://assets.your-domain.com"
  storage: "s3://your-bucket"
  cache_ttl: 24h

auth:
  provider: "jwt"           # or "oauth", "anonymous"
  oauth:
    providers: [google, github]
```

---

## 🌐 Use Cases

- **Collaborative design** — architects walking clients through buildings
- **Remote education** — instructors teaching in shared 3D spaces
- **VR film production** — directors blocking scenes with remote crew
- **Multi-user training simulations** — medical, industrial, military
- **Social VR communities** — your own VRChat/Rec Room alternative
- **Real-time motion capture** — sharing performances between studios

---

## 🚀 Roadmap

- [x] Core signaling + SFU
- [x] Unity SDK
- [x] Unreal Engine 5 SDK
- [x] WebXR client
- [ ] Avatar system (Ready Player Me integration)
- [ ] Full-body IK over the wire
- [ ] Recording + playback (.vrcap files)
- [ ] Federation (rooms across multiple servers)

---

## 📜 License

Apache-2.0. See [LICENSE](LICENSE).

---

<div align="center">

**Star ⭐ to build your own metaverse.**

</div>
