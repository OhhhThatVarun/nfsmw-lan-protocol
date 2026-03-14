# NFSMW LAN Protocol

Reverse-engineered documentation of the Need for Speed: Most Wanted (2005) LAN server protocol, including packet formats, message flow, and server behavior.

> [!WARNING]
> This project is an independent reverse-engineering and documentation effort. It is **not affiliated with, endorsed by, or associated with Electronic Arts (EA), EA Black Box, or the Need for Speed franchise**.

---

## Overview

This repository documents the reverse-engineered LAN server protocol used by **Need for Speed: Most Wanted (2005)**.

The goal of this project is to provide a clear technical reference for:

- packet formats
- message types
- handshake / discovery flow
- server behavior
- client ↔ server state transitions
- quirks and implementation details observed from original binaries and network captures

This repository is intended for:

- protocol research
- interoperability
- preservation
- custom server implementations
- tooling and experimentation

---

## Project Goals

- Document the original NFSMW LAN server protocol as accurately as possible
- Provide packet-level references for known messages
- Describe protocol flow and state transitions
- Record behavioral observations from the original game and server binaries
- Help others build compatible tools, proxies, or replacement servers

---

## Non-Goals

- Distributing proprietary game files or binaries
- Circumventing DRM or copy protection
- Enabling cheating in public or online services
- Repackaging or redistributing copyrighted EA assets

---
