# NFSMW Authentication

How a client proves its identity and gets assigned to the game host. This happens every time a player clicks to join a lobby — after they've already seen it via UDP discovery.

---

## Wire Format (confirmed from pcap)

Every EA protocol message — both client and server — uses the same framing:

```
<command: 4 chars> <null padding: 7 bytes> <length prefix: 1 byte> <body> <null terminator: 1 byte>
```

- **Command** is always 4 ASCII characters, followed by 7 `\x00` bytes (total 11 bytes)
- **Length prefix** = `len(body) + 13`  
  where 13 = 11 (header) + 1 (null terminator) + 1 (prefix byte itself)
- **Body** ends with a `\x00` null terminator byte
- If length prefix ≥ 256: 2 bytes big-endian instead of 1

Field separators:
- **Server → Client**: fields separated by `\t`
- **Client → Server**: fields separated by `\n`

**`newsbadc`** is a special case: no body, no null terminator, prefix = `0x0c` (12).  
Exact bytes: `newsbadc\x00\x00\x00\x0c` (8-char command + 3 nulls + prefix).

### Examples (from pcap)

```
@dir server response (94 bytes total):
  40 64 69 72  00 00 00 00 00 00 00   5e  ADDR=192.168.1.17\t...\x00
  @  d  i  r   <--- 7 nulls --->    [94]  body

~png server (35 bytes):
  7e 70 6e 67  00 00 00 00 00 00 00   23  REF=2026.3.10-16:54:36\x00
  ~  p  n  g   <--- 7 nulls --->    [35]

newsbadc (12 bytes):
  6e 65 77 73 62 61 64 63  00 00 00  0c
  n  e  w  s  b  a  d  c  <- 3 -> [12]   (no body, no null terminator)
```

---

## Overview

Authentication is a **two-connection process**. The first connection is short-lived and only exists to look up the game host address. The second connection is the main session that stays open for the entire time the player is in the game.

```
Client                     Game Host
  │                               │
  │── TCP connect (port 9900) ───▶│  connection 1: directory lookup
  │── @tic + @dir ──────────────▶ │  (sent in same TCP segment)
  │◀─ @dir (ADDR, PORT, SESS) ──  │
  │◀─ FIN+ACK ──────────────────  │  game host closes connection 1
  │                               │
  │── TCP connect (port 9900) ───▶│  connection 2: main session
  │── addr + skey + news ────────▶│  (all three in same TCP segment)
  │◀─ ~png ─────────────────────  │
  │── ~png (pong) ─────────────▶  │  ← pong sent AFTER receiving skey response
  │◀─ skey ─────────────────────  │
  │◀─ newsbadc ─────────────────  │
  │── auth ───────────────────▶   │
  │◀─ auth ─────────────────────  │
  │── pers ───────────────────▶   │
  │◀─ pers ─────────────────────  │  ← authenticated, session active
  │── sele ───────────────────▶   │  ← first post-auth message
```

**Note:** `addr`, `skey`, and `news` all arrive in the same TCP segment on conn2.  
The game host sends `~png` immediately after receiving `addr`, before processing `skey`.  
The client sends pong only after receiving the `skey` response.

---

## Connection 1 — Directory Lookup

### `@tic` — cipher negotiation

Sent immediately after TCP connect, in the same segment as `@dir`.

```
Client → Game Host:
  @tic\x00\x00\x00\x00\x00\x00\x00\x17RC4+MD5-V2\x00
```

`RC4+MD5-V2` is the cipher suite. The game host doesn't respond — it just notes the cipher and waits for `@dir`.

### `@dir` — directory request / response

Client sends its full identity and asks for the game host address:

```
Client → Game Host:
  @dir\x00\x00\x00\x00\x00\x00\x00\xb8
  REGN=NA\nCLST=194010\nNETV=20\nFROM=US\nLANG=EN
  \nMID=$d0577b8c3ad4\nPROD=nfs-pc-2006
  \nVERS="pc/1.3-Nov 21 2005"\nSLUS=SLUS_21351
  \nSKU=14705\nSDKVERS=3.9.3.0\nBUILDDATE="Oct 19 2005"\n\x00
```

| Field | Value | Meaning |
|-------|-------|---------|
| `REGN` | `NA` | Region (North America) |
| `CLST` | `194010` | Cluster ID |
| `NETV` | `20` | Network protocol version |
| `FROM` | `US` | Country |
| `MID` | varies | Machine ID — active network adapter MAC |
| `PROD` | `nfs-pc-2006` | Product identifier |
| `VERS` | `pc/1.3-Nov 21 2005` | Game version |
| `SKU` | `14705` | Disc region |

Game host responds with its address and a session token:

```
Game Host → Client:
  @dir\x00\x00\x00\x00\x00\x00\x00\x5e
  ADDR=192.168.1.17\tPORT=9900\tSESS=1773161639\tMASK=dc83c931a33f2eaf9ab29c6acb60620d\x00
```

| Field | Meaning |
|-------|---------|
| `ADDR` | IP of the game host. Must be reachable by the client. |
| `PORT` | TCP port — always `9900` |
| `SESS` | Session ID — random large integer, unique per session |
| `MASK` | Session token — 32-char hex string. Not validated by client. |

Game host then immediately sends `FIN+ACK` — connection 1 is done.

---

## Connection 2 — Main Session

The client opens a fresh TCP connection to `ADDR:PORT` from the `@dir` response.

### `addr` + `skey` + `news` — client opening burst

All three arrive in the same TCP segment:

```
Client → Game Host:
  addr\x00\x00\x00\x00\x00\x00\x00\x29
  ADDR=192.168.1.5\nPORT=64383\n\x00

  skey\x00\x00\x00\x00\x00\x00\x00\x28
  SKEY=$5075626c6963204b6579\n\x00       ← "Public Key" in hex — a fixed placeholder

  news\x00\x00\x00\x00\x00\x00\x00\x13
  NAME=7\x00                             ← requests MOTD for region 7
```

`ADDR` is the client's own LAN IP. `PORT` is the local port of this TCP connection.  
`SKEY=$5075626c6963204b6579` is always this value — the game doesn't do real asymmetric crypto.

### `~png` — game host ping

Game host sends immediately after receiving `addr`:

```
Game Host → Client:
  ~png\x00\x00\x00\x00\x00\x00\x00\x23
  REF=2026.3.10-16:54:36\x00
```

### `skey` + `newsbadc` — session key and news response

Game host sends both together after processing `skey` from the client:

```
Game Host → Client:
  skey\x00\x00\x00\x00\x00\x00\x00\x33
  SKEY=$ae477db5f42430de640954966012e55e\x00    ← 32-char hex session key

  newsbadc\x00\x00\x00\x0c                      ← no body, no null
```

The session key is a random 16-byte value used to validate the password in `auth`.

### `~png` pong

Client echoes back the `REF` from the game host's ping, but only AFTER receiving the `skey` response:

```
Client → Game Host:
  ~png\x00\x00\x00\x00\x00\x00\x00\x2b
  REF=2026.3.10-16:54:36\nTIME=1\n\x00
```

`TIME=1` is always present in the pong. `REF` is echoed verbatim.

### `auth` — authentication

```
Client → Game Host:
  auth\x00\x00\x00\x00\x00\x00\x01\x02
  REGN=NA\nCLST=194010\nNETV=20\nFROM=US\nLANG=EN
  \nMID=$d0577b8c3ad4\nPROD=nfs-pc-2006
  \nVERS="pc/1.3-Nov 21 2005"\nSLUS=SLUS_21351
  \nSKU=14705\nSDKVERS=3.9.3.0\nBUILDDATE="Oct 19 2005"
  \nNAME=rog\nREGKEY=\nMAC=$d0577b8c3ad4
  \nPASS="~cH1UHJX>#J|HEA{fV%22p\Qe&cslE_P^"\n\x00
```

Note: the 2-byte length prefix `\x01\x02` = 258. Body > 255 bytes → 2-byte big-endian prefix.

| Field | Meaning |
|-------|---------|
| `NAME` | Player's account name |
| `REGKEY` | Registration key — always empty |
| `MAC` | Network adapter MAC, same as `MID` in `@dir` |
| `PASS` | Password hashed with session key (see below) |

**Password algorithm (EA RC4+MD5-V2):**
```
hash1 = MD5("")                          ← MD5 of empty string
hash2 = MD5(hash1 || skey_bytes)         ← skey_bytes = hex-decoded SKEY value
PASS  = eaEncode(hash2)
```
`eaEncode`: splits 16 bytes into 3-byte groups, encodes each group as 4 printable ASCII chars using the 94-char table (ASCII `0x21`–`0x7e`). Output is always 24 chars, wrapped in quotes in the wire format.

The game host responds with the player's profile:

```
Game Host → Client:
  auth\x00\x00\x00\x00\x00\x00\x00\x50
  NAME=rog\tGTAG=rog\tPERSONAS=rog\tXUID=\tTOS=1\tSHARE=1\tADDR=192.168.1.5\x00
```

| Field | Meaning |
|-------|---------|
| `NAME` | Display name (echoed from client) |
| `GTAG` | Gamertag |
| `PERSONAS` | Persona name(s) available |
| `TOS` | Terms of service accepted (always `1`) |
| `ADDR` | Client's IP as seen by the game host |

**In LAN spoofing mode:** password is not validated — any `PASS` is accepted.

### `pers` — persona selection

Client picks which persona to use:

```
Client → Game Host:
  pers\x00\x00\x00\x00\x00\x00\x00\x2e
  PERS=rog\nMAC=$d0577b8c3ad4\nCDEV=\n\x00
```

`PERS` is chosen from the `PERSONAS` list in the `auth` response. `MAC` is the same adapter MAC as before. `CDEV` is always empty.

Game host responds with the full persona record:

```
Game Host → Client:
  pers\x00\x00\x00\x00\x00\x00\x00\x48
  HNAME=rog\tPERS=rog\tLOC=enUS\tMA=\tA=192.168.1.5\tLA=192.168.1.5\x00
```

| Field | Meaning |
|-------|---------|
| `HNAME` | Handle/display name |
| `PERS` | Persona name |
| `LOC` | Locale (always `enUS`) |
| `MA` | Always empty |
| `A` | Client's IP as seen by the game host |
| `LA` | LAN address. On a real LAN: same as `A` (client IP). |

After `pers` the client is fully authenticated. It immediately sends `sele` to enter the lobby browser — see SESSION.md.

---

## Timing (from real session)

```
TCP SYN+ACK (conn 1)
  @tic + @dir sent
  @dir response + FIN received               ← ~250ms round trip
TCP SYN+ACK (conn 2)
  addr + skey + news sent
  ~png received
  skey + newsbadc received
  ~png pong sent
  auth sent
  auth response received                     ← ~180ms
  pers sent
  pers response received                     ← authenticated, ~1.1s total
  sele sent                                  ← immediately after pers
```
