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
dns_records = {}
ip_addresses = {}
ports = {}
hostname_services = {}
ip_services = {}

def name_exists(name: str) -> bool:
    return all_names.get(name) is not None

def use_name(name: str) -> bool:
    if name_exists(name):
        return False
    all_names[name] = True
    return True

def make_tags() -> list:
    tags = {}
    for i in range(0, random.randint(0,3)):
        tags[random.choice(["alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa", "quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey", "xray", "yankee", "zulu"])] = True
    return list(tags.keys())

def make_uuid() -> str:
    return uuid.uuid4().hex

# Add a domain for the org. Returns false if the domain already exists. `expires` is in
# days.
def add_domain(org: str, name: str, expires: int) -> dict:
    domains[org] = domains.get(org, {})
    id = make_uuid()
    domain = {
        "id": id,
        "name": name,
        "registrar": random.choice(regs),
        "registrant_organization": org_names[org],
        "expiry": int(time.time()) + expires * 86400,
        "tags": make_tags(),
    }
    domains[org][id] = domain
    return domain

def add_random_domain(org: str) -> dict:
    name = generate_name() + ".com"
    if not use_name(name): return add_random_domain(org)
    return add_domain(org, name, random.randint(7, 140))

def add_subdomain(org: str, parent_domain: dict, name: str) -> dict:
    id = make_uuid()
    subdomains[org] = subdomains.get(org, {})
    sd = {
        "id": id,
        "parent_id": parent_domain["id"],
        "parent_type": "domain",
        "name": name,
        "tags": make_tags(),
    }
    subdomains[org][id] = sd
    return sd

def add_random_subdomain(org: str) -> dict:
    parent = random.choice(list(domains[org].values()))
    name = generate_name() + "." + parent["name"]
    if not use_name(name): return add_random_subdomain(org)
    return add_subdomain(org, parent, name)

def add_hostname_service(org: str, parent_subdomain: dict, port: str, protocol: str, path: str, ip_list: list, cpe_list: list) -> dict:
    id = make_uuid()
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
        "cpe_list": cpe_list,
        "tags": make_tags(),
    }
    hostname_services[org][id] = hostname_service
    return hostname_service

def add_random_hostname_service(org: str) -> dict:
    parent = add_random_subdomain(org)
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
        ip = add_random_ip_address(org, parent)
        ips.append(ip["ip"])

    cpes = [
        # todo: generate zero or more CPEs
    ]
    # should always be unique since we are creating a subdomain.
    use_name(protocol + "://" + parent["name"] + ":" + str(port) + path)
    hs = add_hostname_service(org, parent, port, protocol, path, ips, cpes)
    return hs

def add_dns_record(org: str, parent_subdomain: dict, record_type: str, value: str) -> dict:
    dns_records[org] = dns_records.get(org, {})
    id = make_uuid()
    record = {
        "id": id,
        "parent_id": parent_subdomain["id"],
        "parent_type": "subdomain",
        "type": record_type,
        "value": value,
        "tags": make_tags(),
    }
    dns_records[org][id] = record
    return record

def add_random_ip_address(org: str, subdomain: dict) -> dict:
    parent = subdomain
    ip_addresses[org] = ip_addresses.get(org, {})
    id = make_uuid()
    ip = f"{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}"
    if not use_name(ip):
        return add_random_ip_address(org)
    dns_record = add_dns_record(org, parent, "A", ip)
    ip_address = {
        "id": id,
        "ip": ip,
        "location": "US",
        "parent_id": dns_record["id"],
        "parent_type": "dns_record",
        "tags": make_tags(),
    }
    ip_addresses[org][id] = ip_address
    return ip_address

def add_port(org: str, parent_ip: dict, port: int, reachable: bool) -> dict:
    ports[org] = ports.get(org, {})
    port_asset = {
        "id": make_uuid(),
        "parent_id": parent_ip["id"],
        "parent_type": "ip_address",
        "protocol": "tcp",
        "port": port,
        "reachable": reachable,
        "tags": make_tags(),
    }
    ports[org][id] = port_asset
    return port_asset

def add_ip_service(org: str, parent_ip: dict, protocol: str, port: int, reachable: bool) -> dict:
    port = add_port(org, parent_ip, port, reachable)
    ip_services[org] = ip_services.get(org, {})
    ip_service = {
        "id": make_uuid(),
        "parent_id": parent_ip["id"],
        "parent_type": "ip_address",
        "hostname": parent_ip["ip"],
        "protocol": protocol,
        "path": "/",
        "port": port,
        "ip_list": [parent_ip["ip"]],
        "cpe_list": [],
        "tags": make_tags(),
    }
    ip_services[org][id] = ip_service
    return ip_service

def add_random_ip_service(org: str) -> dict:
    parent_subdomain = add_random_subdomain(org)
    parent_ip = add_random_ip_address(org, parent_subdomain)
    reachable = True
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

    ip_service = add_ip_service(org, parent_ip, protocol, port, reachable)
    return ip_service

for org in orgs:
    # generate base domains
    for i in range(25):
        add_random_domain(org)
        
    # generate hostname services
    for i in range(200):
        add_random_hostname_service(org)

    for i in subdomains[org].values():
        add_random_ip_address(org, i)

    for i in range(200):
        add_random_ip_service(org)

    # for i in range(200):
    #     parent = random.choice(domains[org])
    #     ips = []
    #     for j in range(0, random.randint(1,3)):
    #         ips.append(f"{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}.{random.randint(0,255)}")
    #     details = {
    #         "class": random.choice(["hostname"]),
    #         "hostname": "www." + parent[1],
    #         "port": random.randint(1, 65535),
    #         "protocol": random.choice(["http", "https", "ftp", "ssh", "smtp"]),
    #         "path": "/",
    #         "ip_list": ips,
    #         "cpe_list": []
    #     }
    #     id = make_uuid()
    #     while all_names.get(details["hostname"]):
    #         details["hostname"] = generate_name() + "." + parent[1] + ".com"
    #     all_names[details["hostname"]] = True
    #     assets.append((id, org, "service", parent[0], "subdomain", json.dumps(details)))


# for i in range(500):
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

with open("config/2.fixtures.sql", "w") as f:
    f.write("-- generated assets --\n")

    std_keys = {"id", "parent_id", "parent_type", "tags"}

    def quote_or_null(value):
        if not value:
            return "NULL"
        return f"'{value}'"

    def write_asset(org: str, asset: dict, asset_type: str):
        details = {key: value for key, value in asset.items() if key not in std_keys}
        f.write(f"INSERT INTO assets (id, org_id, type, parent_id, parent_type, details) VALUES ('{asset['id']}', '{org}', '{asset_type}', {quote_or_null(asset.get('parent_id'))}, {quote_or_null(asset.get('parent_type'))}, '{json.dumps(details)}');\n")

    for org in domains.keys():
        f.write("-- assets for org " + org + " --\n")
        for domain in domains[org].values():
            write_asset(org, domain, "domain")
        for subdomain in subdomains[org].values():
            write_asset(org, subdomain, "subdomain")
        for record in dns_records[org].values():
            write_asset(org, record, "dns_record")
        for ip in ip_addresses[org].values():
            write_asset(org, ip, "ip_address")
        for port in ports[org].values():
            write_asset(org, port, "port")
        for hs in hostname_services[org].values():
            write_asset(org, hs, "hostname_service")
        for ips in ip_services[org].values():
            write_asset(org, ips, "ip_service")

    #for asset in assets:
    #    f.write(f"INSERT INTO assets (id, org_id, type, parent_id, parent_type, details) VALUES ('{asset[0]}', '{asset[1]}', '{asset[2]}', {asset[3] and f"'{asset[3]}'" or 'NULL'}, {asset[4] and f"'{asset[4]}'" or 'NULL'}, '{asset[5]}');\n")
