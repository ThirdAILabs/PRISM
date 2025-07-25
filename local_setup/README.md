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

  7. Integrate Custom Theme in Login UI
    1. Copy the custom-theme folder from the keycloak directory in this repo.
    2. Navigate to the themes folder inside your keycloak-26.0.0 directory.
    3. Paste the directory (named custom-theme) into the themes folder.

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
  3. Make a copy of `cmd/backend/.env.example`.
  ```bash
  cp prism/cmd/backend/.env.example prism/cmd/backend/.env
  ```
  4. Fill in the `.env` file

      a. If using the keycloak setup described above, configure the keycloak args in the config file based on your hosting environment:
  
  <details style="margin-left: 50px;">
    <summary>For local setup</summary>
    
```bash
KEYCLOAK_SERVER_URL="http://localhost/keycloak"
KEYCLOAK_ADMIN_USERNAME="temp_admin"
KEYCLOAK_ADMIN_PASSWORD="password"
KEYCLOAK_PUBLIC_HOSTNAME="http://localhost"
KEYCLOAK_PRIVATE_HOSTNAME="http://localhost"
```
  </details>
  <details style="margin-left: 50px;">
    <summary>For hosted setup (replace example.com with your domain or IP):</summary>
    
```bash
KEYCLOAK_SERVER_URL="http://example.com/keycloak"
KEYCLOAK_ADMIN_USERNAME="temp_admin"
KEYCLOAK_ADMIN_PASSWORD="password"
KEYCLOAK_PUBLIC_HOSTNAME="http://example.com"
KEYCLOAK_PRIVATE_HOSTNAME="http://example.com"
```
      
  </details>
  <br>
  <div style="margin-left: 40px;">
    b. <strong>Rest of the config</strong>
    
```bash
DB_URI="postgresql://<username>:<password>@<host | localhost>:<port | 5432>/prism"
SEARCHABLE_ENTITIES_DATA="<path to PRISM/data/searchable_entities.json>"
# License for PRISM, this should be a keygen license with the Full Access and Prism entitlements.
PRISM_LICENSE="Prism license key"
```
  </div>

5. For Entity search to work, we need to set the openai key as env variable before starting the backend.

```bash
export OPENAI_API_KEY=YOUR_OPENAI_KEY
```

5. Start the backend:

```bash
go run cmd/backend/main.go --env "./cmd/backend/.env"
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

  The worker needs the fund code triangulation database. The database can be created with the command ```psql -U postgres``` followed by ```create database prism_triangulation;```. To populate the database, the following command should be run in the terminal ```pg_restore --no-owner -U postgres -d prism_triangulation -F c prism_triangulation.dump```. The dump can be found [here](https://thirdai-corp-public.s3.us-east-2.amazonaws.com/Prism/prism_triangulation.dump).

  1. Make a copy of `cmd/worker/.env.example` and fill in the fields.
  ```bash
  cp cmd/worker/.env.example cmd/worker/.env
  ```

  2. update the worker config `cmd/worker/.env`:
  ```bash
# Uri for prism postgres db
DB_URI="postgresql://<username>:<password>@<host | localhost>:<port | 5432>/prism"

# Uri for fund code triangulation postgres db
FUNDCODE_TRIANGULATION_DB_URI="postgresql://<username>:<password>@<host | localhost>:<port | 5432>/prism_triangulation"

# License for PRISM, this should be a keygen license with the Full Access and Prism entitlements.
PRISM_LICENSE="prism license key"

# Work dir for worker, will store ndbs and caches etc.
WORK_DIR="any empty directory"

# Path to load data to construct ndbs for author flaggers(update the following path from prism/data)
UNIVERSITY_DATA="<path to PRISM/data/university_webpages.json>"
DOC_DATA="<path to PRISM/data/doc_and_press_releases.json>"
AUX_DATA="<path to PRISM/data/auxiliary_webpages.json>"

# Endpoint for grobid
GROBID_ENDPOINT="http://localhost:8070/" # for local setup
  ```

  3. Start the worker:

  ```bash
  go run cmd/worker/main.go --env "./cmd/worker/.env"
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