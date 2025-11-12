# Cosmos Emerging Threats

When new [Common Vulnerabilities and Exposures (CVE)](https://cve.mitre.org/) are disclosed in popular software, it is usually a race by security teams to determine the impact to their attack surface before attackers weaponize an exploit and use against a vulnerable endpoint.

As this impact analysis can be extremely time-consuming, our customers rely on our **[attack surface management (ASM)](https://bishopfox.com/services/attack-surface-management)** managed service to handle the heavy lifting when it comes to analysis of new CVEs via:

- Threat intelligence
- Asset identification
- Fingerprinting
- Exploit development
- Exploitation

If our team determines a CVE meets our reporting threshold and impacts customer attack surfaces, it is classified as an **emerging threat (ET)**, and workflows are executed to notify customers of any affected assets.

Since 2019, we’ve found that ET execution requires an **accelerated pace** compared to our normal investigative workflow, and that a head start on situational awareness for high-profile CVEs is essential for success.

Given the sheer number of newly disclosed CVEs (nearly **40,000 in 2024**), we found that any strategy to distill these down to an actionable list could greatly improve our reaction time and the speed at which we notify our customers. Out of this need, the **Threat Enablement and Analysis (TEA) team** was formed and tasked with monitoring the constant flow of newly disclosed CVEs—assigning priority to each and determining impact to our customer attack surfaces.

---

## Why CVSS Isn't Enough

The **[Common Vulnerability Scoring System (CVSS)](https://www.first.org/cvss/)** is the de facto standard used to rate the severity of vulnerabilities. While it can be a helpful part of prioritization for ETs, it doesn't tell the whole story. Other attributes vital to real-world risk aren’t necessarily captured by CVSS.

**Examples:**

- An **unauthenticated RCE** in a web application technology with zero instances exposed externally could have a CVSS of **9.9**, but since its use isn’t widespread, the likelihood of impact is nonexistent.
- A CVE with a **high CVSS in common software** may require a specific _non-default configuration_, eliminating the likelihood of a vulnerable instance meeting prerequisites for exploitation.

To account for these gaps, **TEA developed a system of prioritization** based on attributes that provide a more holistic view of real-world impact.

---

## TEA-er’d Prioritization

To separate signal from noise, our team designed a **tiered scoring system (1–3)** that leverages different attributes to rank CVEs by importance.

This system allows us to:

- Immediately rule out CVEs with disqualifying attributes.
- Elevate priority for CVEs with attributes reflecting **real-world impact** to customers.

## Tiers of Prioritization

| **Tier** | **Classification** | **Description**                               |
| -------- | ------------------ | --------------------------------------------- |
| **1**    | Critical Threat    | Imminent impact to customer attack surfaces.  |
| **2**    | Probable Threat    | Potential impact to customer attack surfaces. |
| **3**    | Low Threat         | Unlikely impact to customer attack surfaces.  |

# Cosmos Support Emerging Threat Tiers

Cosmos Product Support categorizes emerging threats by severity to ensure your attack surface's security. The emerging threat tier descriptions below explain the three tier classifications.

<br>
<img src="/markdown-assets/emerging_threat_tiers_v2.png" alt="The three emerging threat tiers with brief description" width="900"/>

## Tier 1 - Critical Threat

Tier 1 threats are critical and require immediate attention. These threats are handled with utmost care and follow a standard procedure. The Cosmos team identifies assets that may be vulnerable, conducts thorough testing, confirms the vulnerabilities, and reports the findings to you promptly.

### Tier 1 Alerts Example and Types

**Example:** Unauthenticated Remote Code Execution in Atlassian Confluence with public proof-of-concept exploit

This example represents a Tier 1 threat, where an attacker can remotely execute code without authentication in Atlassian Confluence. There is public proof-of-concept available, indicating the seriousness of the threat.

**Customer Alerts:**

| **Alert Type**                  | **Alert Content**                                                                                                                       |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| Initial Notification            | Bishop Fox sends you an initial notification regarding the identified threat.                                                           |
| Findings                        | Bishop Fox shares detailed findings from our testing activities.                                                                        |
| Testing Activities              | Bishop Fox updates you on the ongoing testing activities related to the threat.                                                         |
| Follow-up/Wrap-up Notifications | Bishop Fox provides follow-up notifications to ensure you are aware of the latest developments and the conclusion of our investigation. |

### Tier 1 Threat Criteria

To be classified as a Tier 1 threat, a vulnerability must meet one or both of the following criteria:

| **Criteria**                                | **Description**                                                                                                |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Imminent Impact to Customer Environment     | The vulnerability poses an immediate threat to your environment, potentially resulting in severe consequences. |
| Proof-of-Concept (PoC) Exploit Availability | A public PoC exploit exists, indicating that threat actors could easily exploit the vulnerability.             |

_Additional Details:_

| **Criteria**        | **Description**                                                                                                                                                                |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CVSS Score          | Tier 1 threats typically have a CVSS (Common Vulnerability Scoring System) score ranging from 8.0 to 10.0. This score represents the severity and impact of the vulnerability. |
| Attack Vector       | Tier 1 threats generally exploit vulnerabilities through network-based attacks.                                                                                                |
| Complexity          | The complexity level for Tier 1 threats is considered low, meaning that exploiting the vulnerability does not require advanced technical skills.                               |
| Privileges Required | Tier 1 threats can be executed without the need for authentication.                                                                                                            |
| Exploit Maturity    | Tier 1 threats have either a proven exploit or a proof-of-concept available, indicating that attackers can leverage existing techniques to exploit the vulnerability.          |
| Remediation         | Bishop Fox recommends following official, temporary, or vendor-provided workaround solutions to mitigate the Tier 1 threats effectively.                                       |

## Tier 2 - Serious or Probable Threat

Tier 2 threats are considered high and potentially critical, although their validation process differs from Tier 1 threats.

In cases where PoCs are unavailable, authentication is required, or testing on production assets is deemed too risky, an alternative approach is taken. Bishop Fox identifies services that match a specific fingerprint, indicating they are likely vulnerable, and promptly alert you to these potential threats.

### Tier 2 Alerts Example and Types

**Example:** SSRF on SAP Netweaver with no public proof-of-concept exploit

An example of a Tier 2 threat is the Server-Side Request Forgery (SSRF) vulnerability found in SAP Netweaver. While a public proof-of-concept exploit is not available, the vulnerability has been published, and its potential impact on your environment is significant.

**Customer Alerts:**

| **Alert Type**                  | **Alert Content**                                                                                                                     |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Initial Notification            | Bishop Fox sends you an initial notification informing you about the identified threat.                                               |
| Emerging Threat Candidates      | Bishop Fox notifies you about services on your attack surface that match the specific fingerprint indicating potential vulnerability. |
| Testing Activities              | Bishop Fox keeps you updated on ongoing testing activities related to the threat.                                                     |
| Follow-up/Wrap-up Notifications | Bishop Fox provides follow-up notifications to keep you informed of the latest developments and conclude our investigation.           |

### Tier 2 Threat Criteria

To be classified as a Tier 2 threat, a vulnerability must meet the following criteria:

| **Criteria**              | **Description**                                                                                                      |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Vulnerability Publication | The vulnerability has been published, indicating its existence and potential impact.                                 |
| Exploit Availability      | While there may or may not be an available exploit, Bishop Fox continues to monitor and assess the threat landscape. |

_Additional Details:_

| **Criteria**        | **Description**                                                                                                                                                |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CVSS Score          | Tier 2 threats typically have a CVSS score ranging from 7.0 to 10.0, reflecting their severity.                                                                |
| Attack Vector       | Similar to Tier 1 threats, Tier 2 threats primarily target network-based attacks.                                                                              |
| Complexity          | The complexity of Tier 2 threats can range from low to high, meaning that the technical skills required to exploit these vulnerabilities may vary.             |
| Privileges Required | Tier 2 threats can be executed with both unauthenticated and authenticated privileges, although Bishop Fox focuses less frequently on authenticated scenarios. |
| Exploit Maturity    | The exploit maturity for Tier 2 threats can vary. It may be unavailable, or there could be a proof-of-concept or proven exploit available.                     |
| Remediation         | Bishop Fox recommends following official, temporary, workaround, or unofficial solutions to mitigate Tier 2 threats effectively.                               |

## Tier 3 - Low Threat

In addition to Tier 1 and Tier 2 threats, Bishop Fox also monitors and assesses Tier 3 threats. These threats are relatively new and considered to be of lower severity. While they do not merit individual alerts, Bishop Fox regularly analyzes and tracks them.

Tier 3 threats are characterized by their lesser impact on your environment. Although they may not pose an immediate risk, Bishop Fox believes it is essential to keep you informed about their existence.

### Tier 3 Alerts Example and Types

**Example:** Uncommon Wordpress plugin vulnerabilities

An example of a Tier 3 threat includes vulnerabilities found in uncommon Wordpress plugins. While these vulnerabilities may not have widespread exploits or extensive known impacts, Bishop Fox wants to ensure you are aware of their presence.

**Customer Alerts:**

To provide a comprehensive overview of Tier 3 threats, individual alerts are not sent. Instead, Bishop Fox compiles and shares combined reports on a monthly or quarterly basis. These reports help you stay informed about the evolving threat landscape.

### Tier 3 Threat Criteria

To be classified as a Tier 3 threat, a vulnerability must meet the following criteria:

| **Criteria**                                 | **Description**                                                                                                                            |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Low/No Threat to the Customer Attack Surface | These vulnerabilities have limited impact on your overall attack surface.                                                                  |
| CVSS Score                                   | Tier 3 threats typically have a CVSS score ranging from 7.0 to 10.0, indicating their moderate to high severity.                           |
| Attack Vector                                | Similar to Tier 1 and Tier 2 threats, Tier 3 threats primarily target network-based attacks.                                               |
| Complexity                                   | The complexity of Tier 3 threats can vary from low to high, depending on the nature of the vulnerability.                                  |
| Privileges Required                          | Tier 3 threats can be executed with both unauthenticated and authenticated privileges.                                                     |
| Exploit Maturity                             | The exploit maturity for Tier 3 threats can vary. It may be unavailable, or there could be a proof-of-concept or proven exploit available. |
| Remediation                                  | Bishop Fox recommends following official solutions to address Tier 3 threats effectively.                                                  |

Our regular reporting on Tier 3 threats ensures that you have visibility into potential risks, even if their severity is relatively lower, and meets our vision of delivering a comprehensive security service.

Please contact your Cosmos Customer Success Manager if you have any questions or require further assistance.
