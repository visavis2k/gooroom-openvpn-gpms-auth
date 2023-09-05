AUTH_SCRIPT_SOURCE=https://github.com/matevzmihalic/auth-script-openvpn/tarball/master

INSTDIR=/usr/bin
BUILD_DIR=build

all: $(BUILD_DIR)/gooroom-openvpn-gpms-auth $(BUILD_DIR)/auth_script.so

auth-script-openvpn:
	mkdir auth-script-openvpn && wget -O auth-script-openvpn.tar.gz -c $(AUTH_SCRIPT_SOURCE) && tar xvfz auth-script-openvpn.tar.gz -C auth-script-openvpn --strip-components 1

$(BUILD_DIR)/auth_script.so: auth-script-openvpn
	mkdir -p $(BUILD_DIR)
	make -C auth-script-openvpn
	mv auth-script-openvpn/auth_script.so $(BUILD_DIR)

$(BUILD_DIR)/gooroom-openvpn-gpms-auth: *.go
	go build -ldflags="-s -w" -o $(BUILD_DIR)/gooroom-openvpn-gpms-auth .

clean:
	rm -rf $(BUILD_DIR)

install: all
	mkdir -p $(INSTDIR)
	cp $(BUILD_DIR)/* $(INSTDIR)
	chmod 755 $(INSTDIR)/*

package: all
	rm -rf gooroom-openvpn-gpms-auth
	mkdir gooroom-openvpn-gpms-auth
	upx --brute $(BUILD_DIR)/gooroom-openvpn-gpms-auth || true
	cp $(BUILD_DIR)/gooroom-openvpn-gpms-auth $(BUILD_DIR)/auth_script.so gooroom-openvpn-gpms-auth
	tar cvzf gooroom-openvpn-gpms-auth.tar.gz gooroom-openvpn-gpms-auth
	rm -rf gooroom-openvpn-gpms-auth
