#Anonymous messaging using mix networks
## Using the Code
To build and test the code you need:

    *   Go 1.9 or later

Before running or testing the code run 
    
    ```shell
        govendor install +local
        govendor test +local
    ```
    
To perform the unit tests run 

    ```shell
        go test ./...
    ```
    
Before first fresh run of the system run 
    
    ```shell
        bash clean.sh
    ```
This removes all log files and database

To run the network, i.e., mixnodes and providers run

    ```shell
        bash run_network.sh
    ```    
This spins up 3 mixnodes and 1 provider

To simulate the clients run

    ```shell
        bash run_clients.sh
    ```     
