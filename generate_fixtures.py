#!/usr/bin/env python3
# This script generates test fixture data for the assets table.
import json
import random, time, uuid

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

all_names = {}
domains = {}
subdomains = {}
hostname_services = {}

def name_exists(name: str) -> bool:
    return all_names.get(name) is not None

def use_name(name: str) -> bool:
    if name_exists(name):
        return False
    all_names[name] = True
    return True

# Add a domain for the org. Returns false if the domain already exists. `expires` is in
# days.
def add_domain(org: str, name: str, expires: int) -> dict:
    domains[org] = domains.get(org, {})
    id = uuid.uuid4()
    domain = {
        "id": id,
        "name": name,
        "registrar": random.choice(regs),
        "registrant_organization": org_names[org],
        "expiry": int(time.time()) + expires * 86400
    }
    domains[org][id] = domain
    return domain

def add_random_domain(org: str) -> dict:
    name = generate_name() + ".com"
    if not use_name(name): return add_random_domain(org)
    return add_domain(org, name, random.randint(7, 140))

def add_subdomain(org: str, parent_domain: dict, name: str) -> dict:
    id = uuid.uuid4()
    subdomains[org] = subdomains.get(org, {})
    sd = {
        "id": id,
        "parent_id": parent_domain["id"],
        "parent_type": "domain",
        "name": name,
    }
    subdomains[org][id] = sd
    return sd

def add_random_subdomain(org: str) -> dict:
    parent = random.choice(domains[org])
    name = generate_name() + "." + parent["name"]
    if not use_name(name): return add_random_subdomain(org)
    return add_subdomain(org, parent, name)

def add_hostname_service(org: str, parent_subdomain: dict, port: str, protocol: str, path: str, ip_list: list, cpe_list: list) -> dict:
    id = uuid.uuid4()
    hostname_services[org] = hostname_services.get(org, {})
    hostname_service = {
        "id": id,
        "parent_id": parent_subdomain["id"],
        "parent_type": "subdomain",
        "hostname": parent_subdomain["name"],
        "port": port,
        "protocol": protocol,
        "path": path,
        "ip_list": ip_list,
        "cpe_list": cpe_list
    }
    hostname_services[org][id] = hostname_service
    return hostname_service

def add_random_hostname_service(org: str) -> dict:
    parent = random.choice(subdomains[org])
    cls = random.randint(0, 100)
    if cls < 50:
        port = 443
        protocol = "https"
    elif cls < 70:
        port = 80
        protocol = "http"
    elif cls < 80:
        port = 25
        protocol = "smtp"
    else:
        port = random.randint(1, 65535)
        protocol = random.choice(["http", "https", "ftp", "ssh", "smtp"])

    path = "/"
    ips = []
    for j in range(0, random.randint(1,3)):
        ips.append(f"{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}")
    cpes = []
    if not use_name(protocol + "://" + parent["name"] + ":" + str(port) + path):
        return add_random_hostname_service(org) # retry/duplicate
    hs = add_hostname_service(org, parent, port, protocol, path, ips, cpes)
    return hs



for org in orgs:
    # generate domains
    domains[org] = domains.get(org, [])
    for i in range(0, 25):
        add_random_domain(org)
        
    # generate subdomains
    for i in range(0, 200):
        add_random_subdomain(org)

    # generate hostname services
    for i in range(0, 200):
        add_random_hostname_service(org)

    for i in range(



    for i in range(0, 200):
        parent = random.choice(domains[org])
        ips = []
        for j in range(0, random.randint(1,3)):
            ips.append(f"{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}")
        details = {
            "class": random.choice(["hostname"]),
            "hostname": "www." + parent[1],
            "port": random.randint(1, 65535),
            "protocol": random.choice(["http", "https", "ftp", "ssh", "smtp"]),
            "path": "/",
            "ip_list": ips,
            "cpe_list": []
        }
        id = uuid.uuid4()
        while all_names.get(details["hostname"]):
            details["hostname"] = generate_name() + "." + parent[1] + ".com"
        all_names[details["hostname"]] = True
        assets.append((id, org, "service", parent[0], "subdomain", json.dumps(details)))

# for i in range(0, 500):
#     org_id = random.choice(orgs)
#     asset_type = random.choice(asset_types)
#     asset_name = f"Asset_{i}_{asset_type}"

#     if asset_type == "domain":
#         details = {
#             "domain": "www." + generate_name() + ".com",
#             "registrar": random.choice(regs),
#             "registrant_organization": org_names[org_id],
#             "expiry": int(time.time()) + random.randint(1, 20) * 7 * 24 * 3600
#         }
#     elif asset_type == "service":
#         details = {
#             "hostname": "www." + generate_name() + ".com",
#             "port": random.randint(1, 65535),
#             "protocol": random.choice(["http", "https", "ftp", "ssh", "smtp"]),
#             "path": "/",
#         }

#     assets.append((org_id, asset_type, json.dumps(details)))

with open("fixtures.sql", "w") as f:
    f.write("-- generated assets --\n")
    for asset in assets:
        f.write(f"INSERT INTO assets (id, org_id, type, parent_id, parent_type, details) VALUES ('{asset[0]}', '{asset[1]}', '{asset[2]}', {asset[3] and f"'{asset[3]}'" or 'NULL'}, {asset[4] and f"'{asset[4]}'" or 'NULL'}, '{asset[5]}');\n")
