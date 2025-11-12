# CASM Knowledge Base on Subdomain Takeovers

This document outlines the following three types of subdomain takeovers and the coverage that CASM (Cosmos Attack Surface Management) provides for them:

- CNAME records
- A records
- Nameservers

Subdomain takeovers can occur whenever a subdomain’s DNS records point to an asset that no longer exists and is therefore available for registration.

## Takeovers Based on CNAME Records

We see CNAME records-based takeovers most frequently with AWS S3 buckets and AWS Elastic Beanstalk web applications. An example scenario is below:

- Developer creates a bucket named mybucket.s3.amazonaws.com

- Developer creates subdomain bucket.example.com with a CNAME record pointing to mybucket.s3.amazonaws.com

- Sometime later, mybucket.s3.amazonaws.com is deleted, but the subdomain bucket’s DNS record still exists

- Now anyone can create mybucket.s3.amazonaws.com and serve content via your subdomain at bucket.example.com

_CASM automatically detects subdomains that have CNAME records pointing to unregistered domains, which can be taken over by registering the referenced domains with any registrar._

Additionally, many vendors and services are affected by CNAME record vulnerabilities.

_CASM automatically detects subdomain takeovers on the following platforms:_

|                |                 |                  |              |
| -------------- | --------------- | ---------------- | ------------ |
| Aftership      | Aha             | Anima            | Appery       |
| Aquila         | AWS S3          | Big Cartel       | Bitbucket    |
| Branch         | Brightcove      | Campaign Monitor | Fastly       |
| FeedPress      | GetResponse     | Ghost Blog       | GitHub Pages |
| Helpjuice      | Help Scout      | Heroku           | Intercom     |
| JetBrains      | Microsoft Azure | ngrok            | Pagely       |
| Pantheon       | Proposify       | Readme.io        | Shopify      |
| SplashThat     | Statuspage      | Strikingly       | Surge.sh     |
| Teamwork       | Tumblr          | Uberflip         | UserVoice    |
| Vend eCommerce | Wishpond        | Worksites        | Zendesk      |

## Takeovers Based on A Records

A takeover based on A records is a frequent problem with subdomains that have A records pointing to cloud assets. For example:

- Developer creates an EC2 instance, to which Amazon assigns an IP address of 1.2.3.4

- Developer creates subdomain asset.example.com with an A record pointing to 1.2.3.4

- Sometime later, the EC2 instance restarts, automatically releases IP 1.2.3.4, and fetches a new IP of 5.6.7.8, but asset.example.com’s DNS record still points to 1.2.3.4

- Now, anyone who creates an EC2 instance can be randomly assigned the IP address 1.2.3.4 and serve content on asset.example.com

_CASM provides opportunistic coverage for subdomains with A records pointing to abandoned AWS EC2 IP addresses._

## Nameserver (NS) Takeovers

Nameserver takeovers can occur when you use the AWS Route 53 service for managing DNS and the associated nameservers do not have corresponding zone files.

Takeovers can occur in scenarios where:

- Administrator deletes hosted zones from Route 53 but forgets to remove the dangling pointer at the domain registrar level

- Attacker runs a script to continuously create zones for the vulnerable domains until they find a zone with a common nameserver
  - Attacker then adds arbitrary records to the zone that directs the vulnerable domains to IP addresses they control, and the common nameserver is updated accordingly

Less frequently, nameserver takeovers can also succeed when a domain has NS records that point to domains that are themselves unregistered.

_CASM currently provides opportunistic coverage for subdomains with NS records pointing to AWS Route 53._
