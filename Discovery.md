# NFSMW LAN Discovery

How Need for Speed Most Wanted finds other players on a LAN.

---

## What is Discovery?

Before two players can race, their games need to find each other. On a real LAN this happens automatically via **UDP broadcast**. The host shouts to the entire network: "I'm hosting a game, here's my port." Every other machine on the LAN hears it and shows the lobby in the browser.

---

## How it works

The host sends a **384-byte UDP broadcast packet** to `255.255.255.255:9999` every ~3 seconds. The game also sends a **probe packet** (see below) alongside each lobby announcement.

---

## Packet structure

### Lobby announcement packet (384 bytes)

Fields are null-terminated strings at fixed byte offsets. All other bytes are `\x00`.

```
Offset  Length  Content
------  ------  -------
0       3       Magic: gEA
3       1       \x03 (version major)
4       1       Version minor — observed \x05, \x06, \x09; ignored by receiver, any value works
5       3       Session identifier — random 3 bytes, changes every game launch (e.g. r\xaf/, eb\xf4\x07)
8       7       NFSMWNA — game identifier (5 bytes) + region suffix (2 bytes), concatenated, no separator
15      25      \x00 padding
40      var     Lobby name, null-terminated (e.g. MyLobby\x00); max 12 characters — longer names are ignored by the receiver
?       var     \x00 padding; adjusts so that offset 72 is always reached
72      7       Player count field, null-terminated (e.g. 9900|1\x00); format: PORT|COUNT
79      179     \x00 padding; adjusts so that offset 258 is always reached
258     var     Capability flags, null-terminated: TCP:~1:1024\tUDP:~1:1024\x00
?       var     \x00 padding to 384 bytes
```

**Total: 384 bytes.**

**Important:** no IP address is embedded in the packet. The receiving machine uses the UDP packet's **source IP from the IP header** to know who is hosting.

### Probe packet (384 bytes)

**Not required** — the receiver shows the lobby without it. Sent by the real game alongside each lobby announcement; purpose unknown.

```
Offset  Content
------  -------
0       gEA\x00\x00  (magic + two null bytes instead of version)
5       \x3f         (observed constant)
6–383   \x00         (all zeros)
```

---

## Sending behaviour

- Source port: a single ephemeral port chosen at game launch (e.g. 57443), kept constant for the session
- Destination: `255.255.255.255:9999`
- Cadence: one pair (probe + lobby) every ~3 seconds

---

## Tested behaviour

| Field | Finding |
|-------|---------|
| Version minor byte (offset 4) | Ignored — any value accepted |
| Game identifier (offset 8) | `NFSMWNA` must match exactly — `NFSMWXX` and `XXXXXXNA` both rejected |
| Lobby name max length | 12 characters — 12+ causes lobby to not appear |
| Player count | Purely cosmetic — any value displays, no filtering observed |
| Probe packet | Not required — lobby appears without it |

---

## On a real LAN

```
Host's PC (192.168.1.10)
  │
  │  UDP → 255.255.255.255:9999
  │  source IP = 192.168.1.10
  │  probe packet + lobby announcement
  │
  ▼
[LAN switch broadcasts to all machines]
  │
  ▼
Client's PC (192.168.1.20)
  sees source IP = 192.168.1.10
  shows lobby "MyLobby" → connects to 192.168.1.10:9900
```

---
