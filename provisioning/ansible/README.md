## ansibleを用いたisucon11-final構築手順

ベンチマークとアプリケーションを両方搭載したサーバーを、合計4台構築する場合の手順の例を記載しています

### 動作確認環境

* 接続先サーバー Ubuntu 20.04
* ホストPC Ubuntu 20.04
* Ansible 2.9.6

### 実行手順

`provisioning/ansible/standalone.hosts`を変更
(ユーザー名とIPアドレスは適宜変更)
```
[standalone]
ubuntu@192.168.11.10
ubuntu@192.168.11.11
ubuntu@192.168.11.12
ubuntu@192.168.11.13
```

`/etc/hosts`と`provisioning/ansible/roles/bench/files/etc/hosts`に以下を追記
(IPアドレスは適宜変更)
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

SSL証明書作成
```
cd provisioning/ansible/roles/contestant/files/etc/nginx/certificates/
mkdir tmp
cd tmp

openssl genrsa -des3 -out server.password.key 2048 # 適当なパスワードを決めて入力
openssl rsa -in server.password.key -out server.key

openssl req -new -key server.key -out server.csr # 何も入力しないでEnterを押し続けてもOK
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt -extfile <(printf "subjectAltName=DNS:*.t.isucon.dev")
```

SSL証明書をホストPCにインストール
(ブラウザから動作確認したい場合のみ)
```
sudo cp server.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

SSL証明書をansible用に配置
```
cp server.crt ../tls-cert.pem
cp server.key ../tls-key.pem
```

ansible実行
(sudoにパスワードが設定されている場合は`--ask-become-pass`オプションを付加する)
```
cd provisioning/ansible
ansible-playbook -i standalone.hosts site.yml
```

サーバーにSSH接続後、以下コマンドを実行することでベンチマークを実行できる
(targetは適宜変更)
```
cd /home/isucon/benchmarker/bin
./benchmarker -target isucholar1.t.isucon.dev -tls 2> /dev/null
```
