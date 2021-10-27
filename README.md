# thanks-bot
Telegram bot for likes and achievements in chats.
## Installation to Digitalocean
### Install the Vault Helm chart

Install the latest version of the Vault Helm chart in HA mode with integrated storage.
```bash
helm install vault hashicorp/vault \
  --set='server.ha.enabled=true' \
  --set='server.ha.raft.enabled=true'
  ```
 ### Initialize and unseal one Vault pod
 
Vault starts [uninitialized](https://www.vaultproject.io/docs/commands/operator/init.html) and in the [sealed](https://www.vaultproject.io/docs/concepts/seal/#why)
state. Prior to initialization the integrated storage backend is not prepared to receive data.

Initialize Vault with one key share and one key threshold.
 ```bash
 kubectl exec vault-0 -- vault operator init -key-shares=1 -key-threshold=1 -format=json > cluster-keys.json
 ```
 Create a variable named `VAULT_UNSEAL_KEY` to capture the Vault unseal key.
 ```bash
 VAULT_UNSEAL_KEY=$(cat cluster-keys.json | jq -r ".unseal_keys_b64[]")
```
After initialization, Vault is configured to know where and how to access the storage, but does not know how to decrypt any of it.
[Unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) is the process of constructing the master key necessary to read the decryption key to decrypt the data, allowing access to the Vault.

Unseal Vault running on the `vault-0` pod.
```bash
kubectl exec vault-0 -- vault operator unseal $VAULT_UNSEAL_KEY
```
### Join the other Vaults to the Vault cluster

Join the Vault server on `vault-1` to the Vault cluster.
```bash
kubectl exec vault-1 -- vault operator raft join http://vault-0.vault-internal:8200
```
This Vault server joins the cluster sealed. To unseal the Vault server requires the same unseal key, `VAULT_UNSEAL_KEY`, provided to the first Vault server.

Unseal the Vault server on `vault-1` with the unseal key.
```bash
kubectl exec vault-1 -- vault operator unseal $VAULT_UNSEAL_KEY
```
The Vault server on `vault-1` is now a functional node within the Vault cluster.

Join the Vault server on `vault-2` to the Vault cluster.
```bash
kubectl exec vault-2 -- vault operator raft join http://vault-0.vault-internal:8200
```
Unseal the Vault server on `vault-2` with the unseal key.
```bash
kubectl exec vault-2 -- vault operator unseal $VAULT_UNSEAL_KEY
```
### Check Raft cluster

Create a variable named `CLUSTER_ROOT_TOKEN` to capture the Vault unseal key.
```bash
CLUSTER_ROOT_TOKEN=$(cat cluster-keys.json | jq -r ".root_token")
```
Login with the root token on the `vault-0` pod.
```bash
kubectl exec vault-0 -- vault login $CLUSTER_ROOT_TOKEN
```
List all the nodes within the Vault cluster for the `vault-0` pod.
```bash
kubectl exec vault-0 -- vault operator raft list-peers
```
This displays all three nodes within the Vault cluster.
### Set a secret in Vault

First, start an interactive shell session on the `vault-0` pod.
```bash
kubectl exec --stdin=true --tty=true vault-0 -- /bin/sh
```
Your system prompt is replaced with a new prompt `/ $`.

Enable `kv-v2` secrets at the path secret.
```bash
vault secrets enable -path=secret kv-v2
```
Create a secret at path `secret/goapp/config` with a `postgresqlURL`.
```bash
vault kv put secret/goapp/config postgresqlURL="postgresql://doadmin:GgoBNuo9t8SRADm1@private-db-postgresql-fra1-do-user-8476558-0.b.db.ondigitalocean.com:25060/defaultdb?sslmode=require"
```
Verify that the secret is defined at the path secret/data/goapp/config.
```bash
vault kv get secret/goapp/config
```
You successfully created the secret for the web application.
### Configure Kubernetes authentication

Enable the Kubernetes authentication method.
```bash
vault auth enable kubernetes
```
Vault accepts this service token from any client within the Kubernetes cluster. During authentication, Vault verifies that the service account token is valid by querying a configured Kubernetes endpoint.

Configure the Kubernetes authentication method to use the service account token, the location of the Kubernetes host, and its certificate.

#### You can validate the issuer name of your Kubernetes cluster using [this method](https://www.vaultproject.io/docs/auth/kubernetes#discovering-the-service-account-issuer).
Here we use for digitalocean.
```bash
vault write auth/kubernetes/config \
        token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
        kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
        kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
        issuer="https://kubernetes.default.svc.cluster.local"
```
The `token_reviewer_jwt` and `kubernetes_ca_cert` are mounted to the container by Kubernetes when it is created. The environment variable `KUBERNETES_PORT_443_TCP_ADDR` is defined and references the internal network address of the Kubernetes host.

For a client of the Vault server to read the secret data defined in the Set a secret in Vault step requires that the read capability be granted for the path `secret/data/goapp/config`.

Write out the policy named `goapp` that enables the read capability for secrets at path `secret/data/goapp/config`
```bash
vault policy write goapp - << EOF 
path "secret/data/goapp/config" {  capabilities = ["read"]}
EOF
```
Create a Kubernetes authentication role, named `goapp`, that connects the Kubernetes service account name and `goapp` policy.
```bash
vault write auth/kubernetes/role/goapp bound_service_account_names=vault bound_service_account_namespaces=default policies=goapp ttl=24h
```
he role connects the Kubernetes service account, `vault`, and namespace, `default`, with the Vault policy, `goapp`. The tokens returned after authentication are valid for 24 hours.

Lastly, exit the `vault-0` pod.
```bash
exit
```
