#!/usr/bin/env python3
# This script generates test fixture data for the assets table. It needs to be executed
# before starting up the service to initialize the config/2.fixtures.sql file.
import json
import random, time, uuid

# 5 organizations to test against.
orgs = [
    "11111111-1111-1111-1111-111111111111",
    "22222222-2222-2222-2222-222222222222",
    "33333333-3333-3333-3333-333333333333",
    "44444444-4444-4444-4444-444444444444",
    "55555555-5555-5555-5555-555555555555",
]

# Names for those organizations (not used currently).
org_names = {
    "11111111-1111-1111-1111-111111111111": "Alpha Corp",
    "22222222-2222-2222-2222-222222222222": "Beta LLC",
    "33333333-3333-3333-3333-333333333333": "Gamma Inc", 
    "44444444-4444-4444-4444-444444444444": "Delta Ltd",
    "55555555-5555-5555-5555-555555555555": "Epsilon GmbH",
}

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

# Returns true if the given name (global) has been used already.
def name_exists(name: str) -> bool:
    return all_names.get(name) is not None

# Returns true if the given name (global) can be used, and marks it as used.
def use_name(name: str) -> bool:
    if name_exists(name):
        return False
    all_names[name] = True
    return True

# Returns a list of random tags for an asset.
def make_tags() -> list:
    tags = {}
    for i in range(0, random.randint(0,3)):
        tags[random.choice(["alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa", "quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey", "xray", "yankee", "zulu"])] = True
    return list(tags.keys())

# Returns a random UUID hex string.
def make_uuid() -> str:
    return uuid.uuid4().hex

# Add a domain asset for the org and return the new record.
def add_domain(org: str, name: str, expires: int) -> dict:
    if not use_name(name):
        raise Exception("domain name already exists: " + name)
    domains[org] = domains.get(org, {})
    id = make_uuid()
    expiry_time = int(time.time()) + expires * 86400
    domain = {
        "id": id,
        "name": name,
        "registrar": random.choice(regs),
        #"registrant_organization": org_names[org],
        "expiry": expiry_time,
        "expiry_text": time.strftime("%Y-%m-%d", time.gmtime(expiry_time)),
        "tags": make_tags(),
    }
    domains[org][id] = domain
    return domain

# Add a randomly generated domain asset for the org and return the new record.
def add_random_domain(org: str) -> dict:
    name = generate_name() + ".com"
    if not use_name(name): return add_random_domain(org)
    return add_domain(org, name, random.randint(7, 140))

# Add a subdomain asset for the org and return the new record.
def add_subdomain(org: str, parent_domain: dict, name: str) -> dict:
    if not use_name(name):
        raise Exception("subdomain name already exists: " + name)
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

# Add a randomly generated subdomain asset for the org and return the new record.
def add_random_subdomain(org: str) -> dict:
    parent = random.choice(list(domains[org].values()))
    name = generate_name() + "." + parent["name"]
    if not use_name(name): return add_random_subdomain(org)
    return add_subdomain(org, parent, name)

# Add a hostname service asset for the org and return the new record.
def add_hostname_service(org: str, parent_subdomain: dict, port: str, protocol: str, path: str, ip_list: list, cpe_list: list) -> dict:
    used_name = protocol + "://" + parent_subdomain["name"] + ":" + str(port) + path
    if not use_name(used_name):
        raise Exception("hostname service name already exists: " + used_name)
    
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

# Add a randomly generated hostname service asset for the org and return the new record.
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

# Add a DNS record asset for the org and return the new record.
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

# Add a randomly generated IP address asset for the org and return the new record.
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

# Add a port asset for the org and return the new record.
def add_port(org: str, parent_ip: dict, port: int, reachable: bool) -> dict:
    ports[org] = ports.get(org, {})
    id = make_uuid()
    port_asset = {
        "id": id,
        "parent_id": parent_ip["id"],
        "parent_type": "ip_address",
        "protocol": "tcp",
        "port": port,
        "reachable": reachable,
        "tags": make_tags(),
    }
    ports[org][id] = port_asset
    return port_asset

# Add an IP service asset for the org and return the new record.
def add_ip_service(org: str, parent_ip: dict, protocol: str, port: int, reachable: bool) -> dict:
    port = add_port(org, parent_ip, port, reachable)
    ip_services[org] = ip_services.get(org, {})
    id = make_uuid()
    ip_service = {
        "id": id,
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

# Add a randomly generated IP service asset for the org and return the new record. This
# also creates a random subdomain, IP, and corresponding port.
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

# Generate data for each org.
for org in orgs:
    # Add some domains to work with. These will be parents for other resources.
    for i in range(25):
        add_random_domain(org)
        
    # Add hostname services. These will also generate subdomains and IPs.
    for i in range(200):
        add_random_hostname_service(org)

    # Add additional IP addresses for our subdomains.
    for i in subdomains[org].values():
        add_random_ip_address(org, i)

    # Add IP services, which will also generate subdomains and ports.
    for i in range(200):
        add_random_ip_service(org)

# Output to the fixtures SQL file which is loaded during docker compose startup. The
# config folder is a Postgres init folder.
with open("config/2.fixtures.sql", "w") as f:
    f.write("-- generated assets --\n")

    std_keys = {"id", "parent_id", "parent_type", "tags"}

    def quote_or_null(value):
        if not value:
            return "NULL"
        return f"'{value}'"

    def write_asset(content: list, org: str, asset: dict, asset_type: str):
        details = {key: value for key, value in asset.items() if key not in std_keys}
        tags = ["'" + tag + "'" for tag in asset.get("tags", [])]
        tags = ",".join(tags)
        content.append(f"    ('{asset['id']}', '{org}', '{asset_type}', {quote_or_null(asset.get('parent_id'))}, {quote_or_null(asset.get('parent_type'))}, '{json.dumps(details)}', ARRAY[{tags}]::TEXT[])")

    for org in domains.keys():
        f.write("-- assets for org " + org + " --\n")
        f.write(f"INSERT INTO assets (id, org_id, type, parent_id, parent_type, details, tags) VALUES\n")

        print("counts for org", org)
        print("  domains:", len(domains[org]))
        print("  subdomains:", len(subdomains[org]))
        print("  dns_records:", len(dns_records[org]))
        print("  ip_addresses:", len(ip_addresses[org]))
        print("  ports:", len(ports[org]))
        print("  hostname_services:", len(hostname_services[org]))
        print("  ip_services:", len(ip_services[org]))
        content = []
        for domain in domains[org].values():
            write_asset(content, org, domain, "domain")
        for subdomain in subdomains[org].values():
            write_asset(content, org, subdomain, "subdomain")
        for record in dns_records[org].values():
            write_asset(content, org, record, "dns_record")
        for ip in ip_addresses[org].values():
            write_asset(content, org, ip, "ip_address")
        for port in ports[org].values():
            write_asset(content, org, port, "port")
        for hs in hostname_services[org].values():
            write_asset(content, org, hs, "service")
        for ips in ip_services[org].values():
            write_asset(content, org, ips, "service")
        f.write(",\n".join(content) + ";\n\n")
