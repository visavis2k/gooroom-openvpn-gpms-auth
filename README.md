# gooroom-openvpn-gpms-auth

구름플랫폼 관리서버(GPMS)를 바탕으로 OpenVPN 인증을 수행합니다.

## 적용 방법

1. 프로젝트를 빌드합니다.
2. 생성된 auth_script.so 파일과 gooroom-openvpn-gpms-auth 파일을 /usr/bin 디렉터리로 이동합니다.
3. /etc/openvpn/server.conf 파일에 다음 내용을 추가해 플러그인을 등록합니다.

```
plugin /usr/bin/auth_script.so /usr/bin/gooroom-openvpn-gpms-auth
```

4. OpenVPN 서버 구성 파일의 예는 다음과 같습니다.

```
port 1194
proto udp
dev tun
user nobody
group nogroup
persist-key
persist-tun
keepalive 10 120
topology subnet
server 10.8.0.0 255.255.255.0
ifconfig-pool-persist ipp.txt
push "dhcp-option DNS 168.126.63.1"
push "dhcp-option DNS 94.140.15.15"
push "redirect-gateway local def1"
dh none
ecdh-curve prime256v1
tls-crypt tls-crypt.key
ca root_cacert.pem
#cert server_AfYEmXlFWqJm5wrQ.crt
cert vpn_servercert.pem
key vpn_server.key
auth SHA256
cipher AES-128-GCM
ncp-ciphers AES-128-GCM
tls-server
tls-version-min 1.2
tls-cipher TLS-ECDHE-ECDSA-WITH-AES-128-GCM-SHA256
client-config-dir /etc/openvpn/ccd
log /var/log/openvpn/openvpn.log
status /var/log/openvpn/status.log
verb 3
plugin /usr/bin/auth_script.so /usr/bin/gooroom-openvpn-gpms-auth
client-cert-not-required
username-as-common-name
auth-nocache
```

