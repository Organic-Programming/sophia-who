# Sophia Who?

> *"Know thyself."*

Sophia is the holon identity manager. She creates, lists, and pins the
civil status of every holon — UUID, name, clade, lineage, and version.

## Commands

```
who new         — create a new holon identity (interactive)
who show <uuid> — display a holon's identity
who list        — list all known holons (local + cached)
who pin <uuid>  — capture version/commit/arch for a holon's binary
```

## Build

```sh
go build -o who ./cmd/who/
```

## Organic Programming

This holon is part of the [Organic Programming](https://github.com/Organic-Programming/seed)
ecosystem. For context, see:

- [Constitution](https://github.com/Organic-Programming/seed/blob/master/AGENT.md) — what a holon is
- [Methodology](https://github.com/Organic-Programming/seed/blob/master/METHODOLOGY.md) — how to develop with holons
- [Terminology](https://github.com/Organic-Programming/seed/blob/master/TERMINOLOGY.md) — glossary of all terms
- [Contributing](https://github.com/Organic-Programming/seed/blob/master/CONTRIBUTING.md) — governance and standards
