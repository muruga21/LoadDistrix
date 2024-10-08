<a id="top"></a>
<h1 align="center">Load Distrix</h1>
<h3 align="center">    
 
Load Distrix is a simple HTTP load balancer written in Go. It distributes incoming HTTP requests across multiple backend servers, ensuring high availability and fault tolerance. The load balancer uses a round-robin algorithm to cycle through the available backend servers, automatically retrying requests on failure and marking unresponsive servers as down.
</h3>
<hr>





## Features :crystal_ball:

- **Round-Robin Load Balancing:** Distributes requests evenly across backend servers.
- **Health Checks:** Continuously monitors the status of backend servers.
- **Automatic Retry:** Retries requests up to three times in case of failure.
- **Dynamic Server Pool:** Automatically adjusts to backend server availability.

## Requirements

- Go 1.16 or higher
- A configured JSON file named `LoadDistrix.config.json` with backend server details.

## Usage


1. Clone the repository:

   ```sh
   git clone https://github.com/muruga21/LoadDistrix.git
   cd loaddistrix
   ```

2. Prepare the configuration file:

   write loadbalancer config file in the root directory with the following structure:

   ```json
   {
     "backend": [
       {
         "host": "server1",
         "url": "http://127.0.0.1:8081"
       },
       {
         "host": "server2",
         "url": "http://127.0.0.1:8082"
       }
     ]
     ...continue with your backend configuration
   }
   ```

   ```toml
   [backend]
   host="server1"
   url="http://127.0.0.1:8081"
   [backend]
   host="server2"
   url="http://127.0.0.1:8082"

   ... continue with your backend configuration
   ```

   ```yaml
   backend:
- host: "server1"
  url: "http://127.0.0.1:8081"
- host: "server2"
  url: "http://127.0.0.1:8082"

  ...continue with your backend configuration

```

3. Build and run the load balancer:

```sh
go build -o loaddistrix main.go
./loaddistrix <config-file>
````

4. The load balancer will start on port 8000. You can now send HTTP requests to `http://localhost:8000`, and they will be distributed across your configured backend servers.

<hr>

 <details>
   <summary><h3>Configuration :mailbox_with_mail:</h3></summary>


To configure loadbalancer, write a configuration file with your choice of file extension. Each backend server should be specified with its `host` and `url`.

Example:

```json
{
  "backend": [
    {
      "host": "server1",
      "url": "http://127.0.0.1:8081"
    },
    {
      "host": "server2",
      "url": "http://127.0.0.1:8082"
    }
  ]
  ...continue with your backend configuration
}
```

```toml
[backend]
host="server1"
url="http://127.0.0.1:8081"
[backend]
host="server2"
url="http://127.0.0.1:8082"

... continue with your backend configuration
```

```yaml
 backend:
- host: "server1"
 url: "http://127.0.0.1:8081"
- host: "server2"
 url: "http://127.0.0.1:8082"

 ...continue with your backend configuration
```


</details>

## :zap: Featured In:


 <div>
    <h2><img src="https://github.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/blob/master/Emojis/Hand%20gestures/Flexed%20Biceps.png?raw=true" width="35" height="35" > Open Source Programs</h2>
  </div>
<table>
   <tr>
      <th>Event Logo</th>
      <th>Event Name</th>
      <th>Event Description</th>
   </tr>
   <tr>
      <td><img src="https://user-images.githubusercontent.com/63473496/153487849-4f094c16-d21c-463e-9971-98a8af7ba372.png" width="200" height="auto" loading="lazy" alt="GSSoC 24"/></td>
      <td>GirlScript Summer of Code 2024</td>
      <td>GirlScript Summer of Code is a three-month-long Open Source Program conducted every summer by GirlScript Foundation. It is an initiative to bring more beginners to Open-Source Software Development.</td>
   </tr>
</table>
<hr>


<h2 align = "center">Our Contributors ❤️</h2>

<a href="https://github.com/muruga21/LoadDistrix/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=muruga21/LoadDistrix" />
</a>
<hr>


<div>
  <h2><img src="https://fonts.gstatic.com/s/e/notoemoji/latest/1f64f_1f3fb/512.webp" width="35" height="35"> Support </h2>
</div>

<div>
  Don't forget to leave a star<img src="https://fonts.gstatic.com/s/e/notoemoji/latest/1f31f/512.webp" width="35" height="30"> for this project!
</div> <br>

<a href="#top" style="position: fixed; bottom: 20px; right: 20px; background-color: black ; color: white; padding: 10px 20px; text-align: center; text-decoration: none; display: inline-block; border-radius: 5px; font-family: Arial; font-size: 16px;">Go to Top</a>

