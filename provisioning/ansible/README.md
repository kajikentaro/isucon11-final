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

openssl genrsa -out server.key
openssl req -new -key server.key -out server.csr -subj "/C=JP/ST=Tokyo/L=Minato/O=Example Company/OU=IT Department/CN=example.com"
openssl x509 -req -days 3650 -signkey server.key -in server.csr -out server.crt -extfile <(printf "subjectAltName=DNS:*.t.isucon.dev")
```

SSL 証明書をホスト PC にインストール
(ブラウザから動作確認したい場合のみ)

```
sudo cp server.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

SSL 証明書を Nginx 用に配置

```
cp server.crt ../roles/contestant/files/etc/nginx/certificates/tls-cert.pem
cp server.key ../roles/contestant/files/etc/nginx/certificates/tls-key.pem
```

SSL 証明書をシステムにインストールする用に配置

```
mkdir -p ../roles/certificates/files/
cp server.crt ../roles/certificates/files/isucon-self-signed.crt
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
