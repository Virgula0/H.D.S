The **Server** is divided into two main components:

- **Backend**
- **Frontend**

---

## **What Does the Backend Do?**

The **backend** performs the following tasks:

1. Initializes a connection with the **database**.
2. Starts a basic `HTTPServer` to expose the **REST API**, potentially creating **seeds/mock data** for testing purposes.
3. Initializes a `gRPC` server to handle communication with **clients**.
4. Initializes a `TCP` server to handle communication with **daemons**.
5. Encapsulates the **core application logic queue**.

---

## **What Does the Frontend Do?**

The **frontend** performs the following tasks:

1. Starts a basic `HTTPServer` and parses **template files** to expose a user interface.
2. Accepts **user inputs** and communicates with the **backend** using `REST API` by performing **HTTP requests**.

---

## **Compile and Run**

> [!IMPORTANT]  
> You need to export the following **environment variables**. Customize them as needed.

### **1. Start Database**

```bash
cd database
docker build -t dp-database .
docker run -d \
    --name dp-database \
    -e MYSQL_RANDOM_ROOT_PASSWORD=yes \
    --restart unless-stopped \
    -p 3306:3306 \
    --health-cmd="mysqladmin ping -h localhost -uagent -pSUPERSECUREUNCRACKABLEPASSWORD" \
    --health-interval=20s \
    --health-retries=10 \
    dp-database
```

---

### **2. Export Environment Variables**

For simplicity you can save these into a `.env` and the `source .env`
Change them accoding to your needs.

```bash
export BACKEND_HOST="0.0.0.0"
export BACKEND_PORT="4747"
export FRONTEND_HOST="0.0.0.0"
export FRONTEND_PORT="4748"
export DB_USER="agent"
export DB_PASSWORD="SUPERSECUREUNCRACKABLEPASSWORD" # This should be changed (remember to change it in database/initialize.sql too)
export DB_HOST="localhost"
export DB_PORT="3306"
export DB_NAME="dp_hashcat"
export ALLOW_REGISTRATIONS="True" # Disable if needed
export DEBUG="True" 
export RESET="True"
export GRPC_URL="0.0.0.0:7777"
export GRPC_TIMEOUT="10s"
export TCP_ADDRESS="0.0.0.0"
export TCP_PORT="4749"
```

---

### **3. Compile and Run the Server**

```bash
cd server
make build
```

### **4. Run the Server**

```bash
./build/server
```

---

After completing these steps, the **server** should be up and running, with both the **frontend** and **backend** components functioning as expected.
Frontend will be available to `FRONTEND_PORT` port value.
