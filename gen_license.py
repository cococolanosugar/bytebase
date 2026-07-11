#!/usr/bin/env python3
"""
Generate a Bytebase enterprise license JWT token.

Usage:
  python gen_license.py --workspace-id <id> [--private-key <path>] [options]

Examples:
  # Minimal — generates an ENTERPRISE license with 9999 seats/instances, 1 year expiry
  python gen_license.py --workspace-id my-workspace-123

  # Full custom
  python gen_license.py \\
    --workspace-id my-workspace-123 \\
    --private-key private.pem \\
    --plan TEAM \\
    --seats 50 \\
    --instances 10 \\
    --expires-days 365 \\
    --org-name "Acme Corp" \\
    --ha \\
"""

import argparse
from datetime import datetime, timedelta, timezone

import jwt


def main():
    parser = argparse.ArgumentParser(
        description="Generate a Bytebase enterprise license JWT token."
    )
    parser.add_argument(
        "--workspace-id",
        required=True,
        help="Bytebase workspace ID (required).",
    )
    parser.add_argument(
        "--private-key",
        default="private.pem",
        help="Path to RSA private key PEM file (default: private.pem).",
    )
    parser.add_argument(
        "--plan",
        default="ENTERPRISE",
        choices=["FREE", "TEAM", "ENTERPRISE"],
        help="License plan (default: ENTERPRISE).",
    )
    parser.add_argument(
        "--seats",
        type=int,
        default=9999,
        help="Number of seats (default: 9999).",
    )
    parser.add_argument(
        "--instances",
        type=int,
        default=9999,
        help="Number of instances (default: 9999).",
    )
    parser.add_argument(
        "--expires-days",
        type=int,
        default=365,
        help="License validity in days from now (default: 365).",
    )
    parser.add_argument(
        "--org-name",
        default="",
        help="Organization name (optional).",
    )
    parser.add_argument(
        "--ha",
        action="store_true",
        help="Enable high-availability flag.",
    )
    parser.add_argument(
        "--trialing",
        action="store_true",
        help="Mark license as trialing.",
    )

    args = parser.parse_args()

    # Read private key
    with open(args.private_key, "rb") as f:
        private_key = f.read()

    # Build claims — mirrors enterprise/license.go Claims struct + CreateLicense
    now = datetime.now(timezone.utc)
    exp = now + timedelta(days=args.expires_days)

    claims = {
        # Standard registered claims
        "iss": "bytebase",
        "aud": "bb.license",
        "sub": args.workspace_id,
        "iat": now,
        "exp": exp,
        # Custom claims (camelCase to match protojson output Go uses)
        "workspaceId": args.workspace_id,
        "plan": args.plan,
        "seat": args.seats,
        "instance": args.instances,
        "instanceCount": args.instances,
        "ha": args.ha,
        "trialing": args.trialing,
        "orgName": args.org_name,
    }

    # Sign with RS256, kid = "v1"
    token = jwt.encode(claims, private_key, algorithm="RS256", headers={"kid": "v1"})

    print(token)


if __name__ == "__main__":
    main()
