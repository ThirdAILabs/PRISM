# Local Setup Instructions

## Setup Keycloak

1. Download Keycloak version 26.0.0 from the official GitHub repository using this link: [Download Keycloak 26.0.0](https://github.com/keycloak/keycloak/releases/download/26.0.0/keycloak-26.0.0.zip).
2. Extract the downloaded `keycloak-26.0.0.zip` file to a directory of your choice.
3. After extraction, you should have a directory named `keycloak-26.0.0`.
4. Open a terminal and navigate to the `keycloak-26.0.0` directory:
```bash
cd keycloak-26.0.0/
```
5. Start the Keycloak server in development mode with the following command:
```bash
bin/kc.sh start-dev --http-port=8180 --debug --bootstrap-admin-username temp_admin --bootstrap-admin-password password
```
6. To view the admin dashboard go to `localhost:8180` in your browser and login with the credentials `temp_admin` and `password`.

7. Integrate Custom Theme in Login UI
    1. Copy the custom-theme folder from keycloak-assets.
    2. Navigate to the themes folder inside your keycloak-26.0.0 directory.
    3. Paste the extracted directory (named custom-theme) into the themes folder.


## Start the Backend
Note: Right now this assumes you are running on an m1 mac. This repo has thirdai libraries for M1 macs stored with it. If running somewhere else follow instructions below to build thirdai libraries.

Note: For macos the wheels assume that you have libomp installed in `/opt/homebrew/opt/libomp/lib/`, which should be the default if you install with homebrew. You will also need to have openssl3 installed at `/opt/homebrew/Cellar/openssl@3/3.4.0/lib/`. This should also be the default if you install with homebrew.

Ensure that you have a postgres database for PRISM to use if you don't already, this is needed for the backend. If you don't have one you can create one with the command `psql -U postgres` followed by `create database prism;`. This only needs to be done once.

1. Clone the PRISM repo:
```bash
git clone https://github.com/ThirdAILabs/PRISM
```
2. Navigate to the backend folder:
```bash
cd PRISM/prism
```
3. Make a copy of `cmd/backend/config_tmp.yaml` and fill in the fields. If using the keycloak setup described above then the keycloak args in the config file should look like this: 
```yaml
keycloak:
    keycloak_server_url: "http://localhost:8180"
    keycloak_admin_username: "temp_admin"
    keycloak_admin_password: "password"
    public_hostname: "http://localhost"
    private_hostname: "http://localhost"
    ssl_login: false
    verbose: false
```
4. Start the backend: 
```bash
go run cmd/backend/main.go --config "./cmd/backend/config.yaml"
```

## Create a Keycloak User
1. Go back to `localhost:8180` and login with the keycloak admin credentials like in step 6 of keycloak setup. 
2. In the top left there should be a dropdown that defaults to `Keycloak master`. Click on it and select `prism-user` this is changing the realm. 
3. Click on `Users` on the left hand side menu. 
4. Click `Add user` and fill in the username field. Then click `Create` at the bottom. 
5. This will display the user details. Click on `Credentials` at the top and then `Set password` then enter a password and save it.
6. Go back to the `Details` menu (next to the `Credentials` menu) and click the `x` next to `Update Password` in the `Required user actions` section. If you don't do this it will say that the account setup is not complete hwen logging on. 
7. The username and password can now be used to login as a user with keycloak.


## Building ThirdAI Libraries (Optional: Should not be needed for M1 mac)

The following is for building the thirdai libraries needed for the neural db and flash bindings. This is an optional step, the repo has libraries built for m1 mac already in it.

1. Clone Universe:
```bash
git clone https://github.com/ThirdAILabs/Universe --recursive
```
2. Navigate into universe:
```bash
cd Universe
```
3. Build the library: 
  
    Note: you can just use `bin/build.py` without the license options if running locally, however this library will not have licensing so be very careful distributing these libraries. 
```
bin/build.py -f THIRDAI_BUILD_LICENSE THIRDAI_CHECK_LICENSE
```
  

4. Copy the libraries below to `PRISM/prism/search/lib/linux_x64` if building on linux or `PRISM/prism/search/lib/macos_arm64` if running on M1 mac (or other mac os as well but this is not tested yet). After this you should have a 4 `.a` libraries in the directory. See the current `search/lib/macos_arm64` as an example of what it should look like. 
  
    Note: if you build Universe without the licensing flags you will not have the `libcryptopp.a` library. You can skip this. In `PRISM/prism/search/search.go` on lines 3 & 4 you may have to delete the part that says `-lssl -lcrypto` on linux and `-L/opt/homebrew/Cellar/openssl@3/3.4.0/lib/ -lssl -lcrypto` for macos.
    - `Universe/build/libthirdai.a`
    - `Universe/build/deps/rocksdb/librocksdb.a`
    - `Universe/build/deps/utf8proc/libutf8proc.a`
    - `Universe/build/deps/cryptopp-cmake/cryptopp/libcryptopp.a`

5. Now you can run the backend as normal. It will compile the bindings automatically using these libraries if they are in the appropriate directory.
