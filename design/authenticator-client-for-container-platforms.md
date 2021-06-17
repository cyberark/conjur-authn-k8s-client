
# Conjur Authenticator for Container Platforms - Secret-Zero

Created: November 21, 2017
Revised: April 24, 2020

## History

- April 24, 2020 - Reviewed the original design and added some notes about the initial implementation. It is worth
  noting that a more thorough review that adds more details about the current implementation would still be valuable
   - this review was not comprehensive, and was only meant to clear up obvious errors.
- November 30, 2017 - Reviewed and addressed some comments.
- November 27, 2017 - Change the login flow so that the server issues the client’s certificate over a side-channel
  such as ssh or kubectl exec. For authentication, use mutual SSL.
- November 21, 2017 - Initial draft, proposing a login flow which used Server-Sent-Events to coordinate between the
  client and server, as well as an authenticate flow in which the client runs a TLS service which the Authenticator
  can use to obtain the client’s certificate.

## Background

The Container Authenticator issues strong identity credentials to containerized applications running on a platform
such as CloudFoundry, Pivotal CloudFoundry, Kubernetes or OpenShift. The strength of the procedure for establishing
the identity of client containers is very important to security-conscious users.  

### Related Reading

[CloudFoundry Diego Instance identity](https://github.com/cloudfoundry/diego-release/blob/develop/docs/instance-identity.md)

## Requirements

The Container Authenticator must:

- Run inside the container platform itself, e.g. as a managed application / Deployment / Pod.
- Be stateless; new clients must be able to authenticate without requiring writing any data to a Conjur master
  or waiting for replication.
- Be horizontally scalable; as many Authenticators as needed may be run in order to authenticate all the clients.
- Be able to run outside of a Conjur HA appliance, in order to alleviate security concerns about having all the
  enterprise secrets physically available (even if encrypted at rest) inside a container platform environment.
- Provide strong authentication via multiple factors ("secret-zero"). At least one of these factors must use a
  different communication channel from the others.
- Support a time-limited access token or other credential which will quickly expire and can be frequently refreshed.
- Protect access tokens against use in unauthorized locations (e.g. stolen and used elsewhere in the network).

And it should:

- Have an implementation as similar as possible across the supported platforms.

## Architecture

```
                                        +------------+
                                        |E. Security |
                          +------------->   Service  |
                          |             |            |
                          |             +------^-----+
                          |                    |
                          |                    |
+----------------------------------------------------------+
|                         |                    |           |
|                         |                    |           |
| +------------+   +------+-------+   +--------+-------+   |
| |B. Container|   |C. Application+-->+D. Authenticator|   |
| |   Platform +-->+   Containers |   +----------------+   |
| |   Engine   |   +--------------+                        |
| +------------+                                           |
|                                                          |
|                                   A: Compute Cluster     |
+----------------------------------------------------------+
```

A. Compute Cluster - A pool of servers or VMs which run a container engine.
B. Container Platform Engine - Schedules application containers across the Compute Cluster.
C. Application Containers - Individual instances of application code which are scheduled onto the Compute Cluster
   by the Platform Engine
D. Authenticator - A platform application whose function is to verify the identity of Application Containers and issue
   them identity credentials. Runs in the compute cluster.
E. Security Service - A service which provides security methods such as role-based access control, permission checks,
   auditing, and access to secrets. May run inside or outside the compute cluster.

## General design

The Authenticator provides two operations: `login` and `authenticate`. `login` is performed once by each client to
obtain a signed certificate. `authenticate` is performed many times by each client in order to obtain a short-lived
access token. Each access token thus obtained is encrypted by the client's public key.

The `login` method resists identity spoofing by installing the signed certificate onto the client using a Container
Platform Engine API.

Once the client has obtained its signed certificate (which is asynchronously injected into the client's environment
by the Authenticator) it uses the certificate to authenticate, thereby obtaining a short-lived access token.

**Note:** in the original design, we also proposed apply IP restrictions to the short-lived access token. This is not
yet implemented. We also originally planned to have the Authenticator encrypt the access token using the client's
private key (and have it implemented this way for Conjur Enterprise v4), but since the `authenticate` request is sent
using mTLS for Conjur Enterprise (formerly DAP) v10+ and Conjur Open Source v1+, we do not encrypt the access token for these versions.

## Description

### Login

The Authenticator performs a function similar to a Certificate Authority but through Container Platform Engine APIs:
verifying a CSR (certificate signing request) and issuing a certificate for it. The client container generates a
CSR with the following attributes:

- Subject Name: Application identifier (e.g. host identity)
- Subject Alternative Name: Container identifier (e.g. SPIFFE identity of the form
  `spiffe://cluster.local/namespace/{app namespace}/podname/{app pod name}`)

The corresponding private key is stored in a memory file on the client. The client sends the CSR to the Authenticator
using the HTTPS method `POST /inject_client_cert`. When the Authenticator receives a CSR, it verifies the CSR attributes with
the Container Platform Engine. Once the Authenticator has verified the validity of the request, it returns a 200 response
to the Client. The Authenticator continues to process the request asynchronously from this point, and the Authenticator
CA signs the CSR using a CA key which it obtains from a Conjur variable and installs the certificate on the client via
asynchronously via Container Platform Engine APIs. The client stores the signed cert on file and the key is only stored
in memory. The client can use the certificate and key as client TLS information on subsequent requests to the Authenticator.

**Note:** The attributes of the application are different in the actual implementation than in the original design, and
we have updated this section to reflect the current implementation. See the document history for info on the original
design.

#### Login flow

**Note:** The login process is typically performed once per Client Container. It may be repeated if the client knows that
its certificate is due to expire.

1. Create an application container
1. Client container generates a certificate private key and stores it in memory
1. Client container generates a CSR
1. Client container sends the CSR to the Authenticator `inject_client_cert` method
1. Authenticator verifies the Subject Name, Subject Alternative Name(s), and any other attributes of the client request.
1. Authenticator signs the certificate
1. Authenticator installs the certificate onto the client using a side-channel (Container Platform Service API) and in
   an asynchronous thread
1. Authenticator responds with “OK” on the original request channel

### Authenticate

The client wants to obtain an access token from the Authenticator. It sends an HTTPS request to `POST /authenticate`.

The Authenticator verifies the client certificate, and verifies the certificate metadata.  The Authenticator issues an
access token and sends the access token back to the client.

**Notes:**
- `authenticate` is performed as often as the Client Container needs. Access tokens issued by `authenticate`
  have a short lifespan (minutes to hours); the default duration is 8 minutes.
- In the original design (and in the Conjur Enterprise v4 implementation) the Conjur Authenticator encrypts the
  access token using the client certificate before returning it, and the client decrypts the access token upon
  receipt. In the Conjur Enterprise (formerly DAP) v10+ and Conjur Open Source v1+ implementations, the access token is returned over the encrypted
  connection using the same process as the default authenticator.

#### Authenticate flow

1. Client sends an `authenticate` request using mutual SSL
1. Authenticator verifies the client certificate on the request
1. Authenticator issues an access token and encrypts it with the client certificate
1. Authenticator responds to the `authenticate` request with the encrypted access token
1. Client decrypts the access token using the certificate key

At this point, the client container can use the access token to make requests to the security service.

The client’s certificate and key can also be used for other purposes, such as mutual SSL within the compute cluster.

## Discussion

### Why use a certificate rather than issuing a new unique secret token to each client container?

Conjur followers are stateless. As a result, they do not have a writeable data store where issued tokens could be
stored for later verification.  By using signed tokens and certificates, we avoid the need for the Security Service
to have any mutable state, which makes it much easier to guarantee uptime and scalability of the Security Service.

### Why does the Authenticator run inside the Container Platform?

The Authenticator needs to use the network and naming services provided by the Container Platform to establish
side-channel connections to the client containers. If the Authenticator ran outside the Container Platform, each
container which wants to authenticate would have to provide an externally addressable routing path, which is not
desirable or practical.

### Isn’t it insecure to run an Authenticator in a Container Platform?

Like any service, the Authenticator must be protected against unauthorized use or access.

The following types of standard methods can be used to ensure the security the Authenticator. For example:

- The Authenticator runs in a privileged namespace.
- Privileged access (exec / ssh) to the Authenticator is disabled.
- The Authenticator is constrained to run on secure hosts.
- The Authenticator’s identity credentials which it uses to communicate with the Security Service can be secured by
  HSM or Cloud Key Management Service (KMS).

In addition, the Authenticator is only able to issue certificates and access tokens to the compute cluster in which it
runs. Therefore, compromising the Authenticator does not compromise any data which is not already available to the
cluster.  

Finally, the Authenticator does not itself store any secrets or other sensitive data aside from its own identity
credential.  

Further guidance on securing the Authenticator will be provided by CyberArk and the Platform vendor.

### Why make the client send a CSR rather than generating the private key and cert in the Authenticator?

It's more secure if the Authenticator never knows the client private keys. It's also better to make the clients
provide entropy rather than making the Authenticator provide the entropy for all the keys.

### If am able to spoof a request, can I fool the Authenticator into logging me in as any arbitrary application?

No. When you login, the Authenticator will install the certificate onto the client container using a side-channel
routed through the orchestration controller. So, you will also need to intercept this request, and you will need
to MITM the SSL between the Authenticator and the orchestration controller.

### If am able to spoof a request, can I fool the Authenticator into authenticating me as any arbitrary application?

No. The injection of certificates into the proper container via the side-channel Container Platform API ensures that
only the valid client receives the signed certificate. Furthermore, the mutual TLS session for authenticating to
retrieve the access token requires the client's SSL key (which is stored only in the client's memory) and a
valid/signed certificate.
