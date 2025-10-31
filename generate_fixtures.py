#!/usr/bin/env python3
# This script generates test fixture data for the assets table.
import json
import random, time

orgs = [
    "11111111-1111-1111-1111-111111111111",
    "22222222-2222-2222-2222-222222222222",
    "33333333-3333-3333-3333-333333333333",
    "44444444-4444-4444-4444-444444444444",
    "55555555-5555-5555-5555-555555555555",
]

org_names = {
    "11111111-1111-1111-1111-111111111111": "Alpha Corp",
    "22222222-2222-2222-2222-222222222222": "Beta LLC",
    "33333333-3333-3333-3333-333333333333": "Gamma Inc", 
    "44444444-4444-4444-4444-444444444444": "Delta Ltd",
    "55555555-5555-5555-5555-555555555555": "Epsilon GmbH",
}

asset_types = ["domain", "service"]

def generate_name():

    adjectives_opinion = ["awesome", "terrible", "fantastic", "mediocre", "excellent", "poor", "great", "bad", "superb", "awful"]
    adjectives_size = ["big", "small" , "tiny", "huge", "massive", "mini", "gigantic", "colossal", "petite", "enormous"]
    adjectives_noun = ["eagle", "tiger", "lion", "shark", "wolf", "panther", "dragon", "phoenix", "griffin", "unicorn", "falcon", "bear", "leopard", "cougar", "jaguar", "crocodile", "alligator", "rhino", "hippo", "buffalo"]

    return f"{random.choice(adjectives_opinion)}{random.choice(adjectives_size)}{random.choice(adjectives_noun)}"

regs = ["GoDaddy", "Namecheap", "Bluehost", "HostGator", "DreamHost", "1&1 IONOS", "Google Domains", "AWS Route 53", "Hover", "Dynadot"]

assets = []

for i in range(0, 500):
    org_id = random.choice(orgs)
    asset_type = random.choice(asset_types)
    asset_name = f"Asset_{i}_{asset_type}"

    if asset_type == "domain":
        details = {
            "domain": "www." + generate_name() + ".com",
            "registrar": random.choice(regs),
            "registrant_organization": org_names[org_id],
            "expiry": int(time.time()) + random.randint(1, 20) * 7 * 24 * 3600
        }
    elif asset_type == "service":
        details = {
            "hostname": "www." + generate_name() + ".com",
            "port": random.randint(1, 65535),
            "protocol": random.choice(["http", "https", "ftp", "ssh", "smtp"]),
            "path": "/",
        }

    assets.append((org_id, asset_type, json.dumps(details)))

with open("fixtures.sql", "w") as f:
    f.write("-- generated assets --\n")
    for asset in assets:
        f.write(f"INSERT INTO assets (org_id, type, details) VALUES ('{asset[0]}', '{asset[1]}', '{asset[2]}');\n")
