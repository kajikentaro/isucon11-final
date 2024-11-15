## Ansible を用いた isucon11-final 構築手順

ベンチマークとアプリケーションを両方搭載したサーバーを、合計 4 台構築する場合の手順の例を記載しています

### 動作確認環境

- 接続先サーバー Ubuntu 20.04
- ホスト PC Ubuntu 20.04
- Ansible 2.9.6

### 実行手順

`provisioning/ansible/standalone.hosts`を変更
(ユーザー名と IP アドレスは適宜変更)

```
[standalone]
ubuntu@192.168.11.10
ubuntu@192.168.11.11
ubuntu@192.168.11.12
ubuntu@192.168.11.13
```

`/etc/hosts`と`provisioning/ansible/roles/bench/files/etc/hosts`に以下を追記
(IP アドレスは適宜変更)

```
192.168.11.10   isucholar0.t.isucon.dev
192.168.11.11   isucholar1.t.isucon.dev
192.168.11.12   isucholar2.t.isucon.dev
192.168.11.13   isucholar3.t.isucon.dev
```

ファイル生成

```
cd provisioning/packer/
make files-generated/REVISION
make files-generated/isucon11-final.tar
make files-generated/benchmarker
cp files-generated /dev/shm -r
```

SSL 証明書作成

```
cd provisioning/ansible/
mkdir -p tmp
cd tmp

# CA用秘密鍵とCA証明書作成
openssl genrsa -out server-CA.key 2048
openssl req -x509 -new -nodes -key server-CA.key -sha256 -days 3650 -out server-CA.crt -subj "/C=JP/ST=Tokyo/L=Minato/O=Example Company/OU=Certificate Authority/CN=example-CA"

# サーバー用秘密鍵とCSR作成
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/C=JP/ST=Tokyo/L=Minato/O=Example Company/OU=IT Department/CN=example.com"

# CA用秘密鍵でサーバー証明書を作成
openssl x509 -req -in server.csr -CA server-CA.crt -CAkey server-CA.key -CAcreateserial -out server.crt -days 3650 -sha256 -extfile <(printf "subjectAltName=DNS:*.t.isucon.dev")
```

CA 証明書をホスト PC にインストール(実際にアプリケーションの動作確認する場合のみ)

- Chrome などのブラウザで確認したい場合はブラウザの設定から`server-CA.crt`をインストールする
- curl コマンド等で確認したい場合は以下のコマンドを実行する

```
sudo cp server-CA.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

サーバー証明書を Nginx 用に配置

```
cp server.crt ../roles/contestant/files/etc/nginx/certificates/tls-cert.pem
cp server.key ../roles/contestant/files/etc/nginx/certificates/tls-key.pem
```

CA 証明書をシステムにインストールする用に配置

```
mkdir -p ../roles/certificates/files/
cp server-CA.crt ../roles/certificates/files/isucon-self-signed.crt
```

ansible 実行
(sudo にパスワードが設定されている場合は`--ask-become-pass`オプションを付加する)

```
cd provisioning/ansible
ansible-playbook -i standalone.hosts site.yml
```

サーバーに SSH 接続後、以下コマンドを実行することでベンチマークを実行できる
(target は適宜変更)

```
cd /home/isucon/benchmarker/bin
./benchmarker -target isucholar1.t.isucon.dev -tls 2> /dev/null
```
