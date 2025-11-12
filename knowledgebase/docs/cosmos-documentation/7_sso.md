# Setting up Single Sign-On (SSO) for Cosmos

Here are instructions to set up single sign on (SSO) with the Cosmos portal. This document outlines the SSO setup process and the key pieces of information needed from you to complete the process. Examples using Okta and Active Directory are provided below.

## SSO Overview

Bishop Fox supports an SSO/SAML integration with the Cosmos portal where Bishop Fox serves as the Service Provider (SP) in an Identity Provider (IdP)-initiated SSO flow.

The SSO setup process is as follows:

- Bishop Fox will provide you information to set up a connector client.
  - [ ] A sign in URL

  <br />
  - [ ] Entity ID

  <br />

- Once that connector client is configured, Bishop Fox needs the following items from you:
  - [ ] An X.509 certificate for the environment

  <br />
  - [ ] A sign in URL (SAML SSO URL, generally unique per environment)

  <br />
  - [ ] A base email domain (typically the @<company_name>.com part of a client’s email)

- Once both items above are completed, a Bishop Fox account manager will schedule a working call to test the SSO functionality.

## SSO Setup Process

The SSO setup process will be unique to your organization and your SSO provider.

Active Directory (AD) SSO setup is provided as an example below. The process will likely be similar, with differences, depending on your SSO provider.

No matter what SSO provider you use, Bishop Fox requires the following attributes:

- email set to user’s email address

- email_verified set to true (not "true")

### Okta SSO Setup Process Example

These instructions walk you through initiating SSO integration with the Cosmos platform and your Okta instance.

It is assumed that you already have Okta Admin level access to the App Integration Wizard (AIW).

### Add Cosmos to Okta Using the App Integration Wizard (AIW)

1. Contact your Bishop Fox Customer Success Manager (CSM) (SDM) to request SSO integration.

2. Bishop Fox will provide you the following information so that you can set up a client connector:

   a. A Sign on URL (unique per environment). In the format of `https://bishopfox.auth0.com/login/callback?connection=<SAML_connector_name>`
   1. All `<SAML_connector_names>` will look like `BF-<client_name>-SAML`. For example, if Bishop Fox was a client, it would be `BF-BISHOPFOX-SAML`.

   b. A service EntityID (unique per environment): `urn:auth0:bishopfox:<SAML_connector_name>`

3. Once you’ve received the Sign in URL and service EntityID, log in to your Okta instance and do the following:

   a. Navigate to the _Applications_ screen and click **Create App Integration**.

<img src="/markdown-assets/create-app-integration.png" alt="Okta sso create app integration screen" width="900"/>

This displays the _Create a new app integration_ dialog box.

<img src="/markdown-assets/create-new-app.png" alt="Okta create new app screen with SAML radio button selected" width="900"/>

b. Click the **SAML 2.0** radio button and click **Next** to display the _Create SAML Integration_ screen.

<img src="/markdown-assets/create-saml-integration.png" alt="Okta create SAML integration screen general settings fields" width="900"/>

c. In the _App name_ field, enter the SAML connector name provided by Bishop Fox in 2.a.1. above. For example, BF-TEST-SAML.

d. Click **Next** to display the _SAML Settings_ fields and enter the values from 2.a and 2.b above in the _Single sign-on URL_ field and the _Audience URI (SP Entity ID)_ field. BF-TEST-SAML is used as an example.

<img src="/markdown-assets/enter-app-name.png" alt="Okta SAML settings fields filled in" width="900"/>

e. Scroll down to the _Attribute Statements (optional)_ field group and enter the following:

1. Under the _Name/Name format_ column, enter “email” in the left field and leave the name format drop-down list default.

2. In the _Value_ column, select “user.email” from the drop-down list.

3. Click **Add Another** to add another row.

4. Under the _Name/Name format_ column, enter “email_verified” in the left field and leave the name format drop-down list default.

5. In the _Value_ column, type “true”.

  <img src="/markdown-assets/attribute-statements.png" alt="Okta attribute statments fields filled in" width="900"/>

f. Click **Next** to display the Okta support screen and enter any values you wish.

g. Click **Finish** to display the SAML summary page.

  <img src="/markdown-assets/feedback.png" alt="Okta SAML setup summary screen" width="900"/>

h. Scroll down to view the _SAML Signing Certificates_ section and download the IdP metadata and X.509 certificate by doing the following:

1. Click **Actions -> View IdP metadata** to display the metadata in a new browser tab.

  <img src="/markdown-assets/idp-metadata.png" alt="Okta SAML signing certificates screen with Actions drop-down list selected showing View IdP metadata option" width="900"/>

Note the Location value near the end of the metadata (in the example above, “https://trial-4463928.okta.com/app/trial-4463928_testapp_1/exk4g9va6hg6JJXZk697/sso/saml”).

  <img src="/markdown-assets/idp-meta-large.png" alt="Okta metadata opened in new browser tab showing the url" width="900"/>

2. Click **Actions -> Download certificate** for the active certificate to download the X.509 certificate.

  <img src="/markdown-assets/download-cert.png" alt="sso SAML card screen in AD" width="900"/>

i. Share the following with Bishop Fox team in a secure manner so that they can complete the SSO setup for your organization:

1. The XML metadata file (or the Location value from it, “https://trial-4463928.okta.com/app/trial-4463928_testapp_1/exk4g9va6hg6JJXZk697/sso/saml” in the example above)

2. The raw X.509 certificate file

3. The Bishop Fox team will follow up with you when setup is complete and then testing can begin.

---

### Active Directory SSO Setup Process Example

These instructions walk you through initiating SSO integration with the Cosmos platform and your AD tenant.

It is assumed that you already have:

- An Azure AD tenant

- Login access as a Global Administrator, Cloud Application Administrator, or Application Administrator.

### Add Cosmos to Your Azure Active Directory (Azure AD) Tenant

1. Contact your Bishop Fox Customer Success Manager (CSM) (SDM) to request SSO integration.

2. Bishop Fox will provide you the following information so that you can set up a client connector:

   a. A Reply URL (Assertion Consumer Service) that is unique per environment; it is in the format of `https://bishopfox.auth0.com/login/callback?connection=<SAML_connector_name>`
   1. All `<SAML_connector_names>` will look like `BF-<client_name>-SAML`. For example, if Bishop Fox was a client, it would be `BF-BISHOPFOX-SAML`

   b. A service EntityID (unique per environment): `urn:auth0:bishopfox:<SAML_connector_name>`

3. Once you’ve received the Sign in URL and service EntityID, log in to your Azure Active Directory and do the following:

   a. Follow the Azure Active Directory instructions to add an enterprise application, naming it something like “Bishop Fox” or “Cosmos”, whichever is clearest for you.

   b. Follow the Azure Active Directory instructions to create and assign a user account for the Cosmos application that you added just above.

   c. From the Dashboard -> Enterprise applications screen, select the Cosmos application (or however you’ve named it in step 2.1., above) to display the Properties screen.

   d. Click Single sign-on in the left hand navigation.

    <img src="/markdown-assets/bf_sso1.png" alt="getting started SSO screen in AD" width="900"/>

   e. Click the SAML card from the Select a single sign-on method screen to display the Set up Single Sign-On with SAML screen.

     <img src="/markdown-assets/bf_sso2.png" alt="sso SAML card screen in AD" width="900"/>

   f. In the SAML-based Sign-on screen, click Edit for the Basic SAML Configuration field group.

     <img src="/markdown-assets/bf_sso3.png" alt="setup single sign on with SAML screen in AD" width="900"/>

   This displays the Basic SAML Configuration dialog box.

   g. In the Basic SAML Configuration dialog box, enter the values from step 2.1. and 2.1. above in the following fields:
   1. Click the Add identifier link to display a field where you can enter the Identifier (Entity ID) value provided by Bishop Fox in Step 2.2 above.

   2. Click the Add reply URL link under the Reply URL (Assertion Consumer Service URL) section to enter the value provided by Bishop Fox in Step 2.1. above.

   <img src="/markdown-assets/bf_sso4.png" alt="basic SAML Configuration screen in AD" width="600"/>
   3. Save your changes.

   h. Complete any required fields in the Attributes & Claims fields (shown in the screen shot in step 3.f. above).

   i. From the SAML Certificates fields, take the following actions and share the value and the certificate in a secure manner with the Bishop Fox team so that they can complete the SSO setup for your organization:
   1. Copy the App Federation Metadata Url value

   2. Downloaded Certificate (Raw) and attach it to the secure communication

   <img src="/markdown-assets/bf_sso6.png" alt="SAML-based Sign On Screen in AD showing SAML Certificates" width="900"/>

4. Client needs to pass the following attributes:

   a. email set to user’s email address

   b. email_verified set to boolean true (not string "true")

   <img src="/markdown-assets/SSO_Attributes_and_Claims.png" alt="SSO Attributes and Claims" width="900"/>

<img src="/markdown-assets/SAML_based_sign_on.png" alt="SAML-based Sign On" width="900"/>
