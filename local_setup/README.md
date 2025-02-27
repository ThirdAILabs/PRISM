# Local Setup Instructions

<details>
  <summary><h2 style="display: inline;">Clone the repo</h2></summary>
  <br>

  Run the commands
  ```bash
  git clone https://github.com/ThirdAILabs/PRISM
  cd PRISM
  ```
</details>
<br>
<details>
  <summary><h2 style="display: inline;">Launch Traefik</h2></summary>
  <br>

  Run the commands
  1. Install Traefik using Homebrew:
  ```bash
  brew install traefik
  ```

  2. Navigate to the local_setup folder in the PRISM repository and run:
  ```bash
  cd local_setup
  bash launch_traefik.sh
  ```
  **Note: Ignore the error about non-empty provider endpoint.**
</details>
<br>
<details>
  <summary><h2 style="display: inline;">Setup Keycloak</h2></summary>
  <br>

  1. Download Keycloak version 26.0.0 from the official GitHub repository using this link: [Download Keycloak 26.0.0](https://thirdai-corp-public.s3.us-east-2.amazonaws.com/keycloak/keycloak-26.0.0.zip).
  2. Extract the downloaded `keycloak-26.0.0.zip` file to a directory of your choice.
  3. After extraction, you should have a directory named `keycloak-26.0.0`.
  4. Open a terminal and navigate to the `keycloak-26.0.0` directory:

  ```bash
  cd keycloak-26.0.0/
  ```

  5. Start the Keycloak server in development mode with the following command:

  ```bash
  bin/kc.sh start-dev --http-port=8180 --debug --bootstrap-admin-username temp_admin --bootstrap-admin-password password --hostname-strict false --proxy-headers forwarded --http-relative-path /keycloak
  ```

  6. To view the admin dashboard go to `localhost:8180` in your browser and login with the credentials `temp_admin` and `password`.
</details>
<br>
<details>
  <summary><h2 style="display: inline;">Building ThirdAI Libraries (Optional: Should not be needed for M1 mac os 15)</h2></summary>
  <br>

  The following is for building the thirdai libraries needed for the neural db and flash bindings. This is an optional step, the repo has libraries built for `m1 mac os 15` already in it.

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
</details>
<br>
<details>
  <summary><h2 style="display: inline;">Start the Backend</h2></summary>
  <br>

  Note: For macos the wheels assume that you have libomp installed in `/opt/homebrew/opt/libomp/lib/`, which should be the default if you install with homebrew. You will also need to have openssl3 installed at `/opt/homebrew/Cellar/openssl@3/3.4.0/lib/`. This should also be the default if you install with homebrew.

  Prism needs a database for working, create one if not already done.

  1. Connect with psql client
  ```bash
  psql -U <username> -d postgres
  ```
  2. Create database
  ```sql
  create database prism;
  ```
  3. Make a copy of `cmd/backend/config_tmp.yaml`.
  ```bash
  cp prism/cmd/backend/config_tmp.yaml prism/cmd/backend/config.yaml
  ```
  4. Fill in the `config.yaml`

      a. If using the keycloak setup described above, configure the keycloak args in the config file based on your hosting environment:
  
  <details style="margin-left: 50px;">
    <summary>For local setup</summary>
    
```yaml
keycloak:
  keycloak_server_url: "http://localhost/keycloak"
  keycloak_admin_username: "temp_admin"
  keycloak_admin_password: "password"
  public_hostname: "http://localhost"
  private_hostname: "http://localhost"
  ssl_login: false
  verbose: false
```
  </details>
  <details style="margin-left: 50px;">
    <summary>For hosted setup (replace example.com with your domain or IP):</summary>
    
```yaml
keycloak:
  keycloak_server_url: "http://example.com/keycloak"
  keycloak_admin_username: "temp_admin"
  keycloak_admin_password: "password"
  public_hostname: "http://example.com"
  private_hostname: "http://example.com"
  ssl_login: false
  verbose: false
```
      
  </details>
  <br>
  <div style="margin-left: 40px;">
    b. <strong>Rest of the config</strong>
    
```yaml
postgres_uri: postgresql://<username>:<password>@<host | localhost>:<port | 5432>/prism
searchable_entities: <path to PRISM/data/searchable_entities.json>
ndb_license: "Bolt license key"
```
  </div>

5. Start the backend:

```bash
go run cmd/backend/main.go --config "./cmd/backend/config.yaml"
```
</details>
<br>

<details>
  <summary><h2 style="display: inline;">Create a Keycloak User</h2></summary>
  <br>

  1. Go to `localhost:8180/keycloak` and log in with the Keycloak admin credentials from step 6 of Keycloak setup.
  2. In the top left, select `prism-user` from the dropdown to change the realm.
  3. Click `Users` on the left-hand menu.
  4. Click `Add user`, fill in the username, email, First Name, Last Name fields, and click `Create` at the bottom.
  5. Go to the `Credentials` tab, click `Set password`, enter a password, and save it.
  6. In the `Details` tab, remove the `Update Password` requirement under `Required User Actions`.
  7. The username and password can now be used to log in as a user with Keycloak.

  <div style="margin-left: 20px;">

  ### **Adding an Admin User in the `prism-admin` Realm**

  Follow the same steps as above, but select the `prism-admin` realm instead of `prism-user`. Create an admin user with credentials that will be used in the Bash script.

  ## Running the License Automation Script

  1. Navigate to the directory where the script is stored:
  ```bash
  cd PRISM/local_setup
  ```

  2. Ensure you have `jq` installed:
  ```bash
  sudo apt install jq  # Ubuntu/Debian
  brew install jq      # macOS
  ```

  3. Run the script:
  ```bash
  ./create_license.sh
  ```

  The script will:

  - Fetch an admin access token from `prism-admin` realm and create a license.
  - Fetch a user access token from `prism-user` realm and activate the license for that user.
  - Print the activation response to confirm success.
  </div>
  
  </details>
<br>

<details>
<summary><h2 style="display: inline;">Setup Grobid</h2></summary>
  <br>

  Grobid can be set up on Blade server and can be accessed by forwarding the port.
  
  Run the command ```docker run --rm --init --ulimit core=0 -p 8070:8070 grobid/grobid:0.8.0```. This will start Grobid on port ```8070```.
</details>
<br>

<details>
  <summary><h2 style="display: inline;">Start the worker</h2></summary>
  <br>

  1. Make a copy of `cmd/worker/config_tmp.yaml` and fill in the fields.
  ```bash
  cp cmd/worker/config_tmp.yaml cmd/worker/config.yaml
  ```

  2. update the worker config `cmd/worker/config.yaml`:
  ```yaml
# Uri for prism postgres db
postgres_uri: "postgresql://<username>:<password>@<host | localhost>:<port | 5432>/prism"

# Logfile for backend
logfile: "./worker.log"

# License for NDB
ndb_license: "bolt license key"

# Work dir for worker, will store ndbs and caches etc.
work_dir: "any empty directory"

# Path to load data to construct ndbs for author flaggers(update the following path from prism/data)
ndb_data:
  university: "<path to PRISM/data/university_webpages.json>"
  doc: "<path to PRISM/data/doc_and_press_releases.json>"
  aux: "<path to PRISM/data/auxiliary_webpages.json>"

# Endpoint for grobid
grobid_endpoint: "http://localhost:8070/" # for local setup
  ```

  3. Start the worker:

  ```bash
  go run cmd/worker/main.go --config "./cmd/worker/config.yaml"
  ```

</details>
<br>
<details>
  <summary><h2 style="display: inline;">Start the Frontend</h2></summary>
  <br>

  1. Navigate to the frontend folder:

  ```bash
  cd PRISM/frontend
  ```

  2. Create and configure the `.env` file:

  __Important Note__: Please ensure that you enter the URL values without quotes and remove any inline comments that might appear on the same line.

  - For local development:
    ```bash
    REACT_APP_API_URL=http://localhost
    REACT_APP_KEYCLOAK_URL=http://localhost/keycloak
    ```

  - For hosted setup (replace example.com with your domain or IP):
    ```bash
    REACT_APP_API_URL=http://example.com
    REACT_APP_KEYCLOAK_URL=http://example.com/keycloak
    ```

  3. Install dependencies:

  ```bash
  npm i
  ```

  4. Start the frontend development server:

  ```bash
  npm start
  ```

  The frontend will be accessible at `http://localhost` in your browser.
</details>