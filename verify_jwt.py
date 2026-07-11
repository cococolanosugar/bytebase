"""Verify a generated JWT license token against the public key."""
import sys
import json
import jwt

token = sys.stdin.read().strip()
with open("public.pem", "rb") as f:
    pubkey = f.read()

claims = jwt.decode(token, pubkey, algorithms=["RS256"], audience="bb.license")
print(json.dumps(claims, indent=2, default=str))
