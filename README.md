# Product-catalog
- this project is simple and can be use as a boilerplat for future projects.

# index

1. developer setup
    - a. postgres
        - i. docker
    - b. backend
        - i. golang and fiber
        - ii. air (watcher)
        - iii. migration (before server run)
        - iv. server (app)
    - c. frontend setup
        - i. node and vue3
        - ii. server (watcher)
2. Tools information
    - a. Backend
        - i. migration
        - ii. orm
        - iii. uuid
        - iv. testing
    - b. Frontend
        - i. material UI
        - ii. testing
    - c. General
        - i. hooks
3. production setup
    - a. config envs
    - b. docker setup
4. scope description
    - a. form simulator
    - b. duplicate order aggrigator (Cli)
5. assumptions
6. project description
7. project structure
    - a. backend
    - b. frontend


## 1. Developer setup

- This project will be monorepo.
- So frontend will be in app and backend will be in server directory.
- Thired we use postgres as database and run inside the docker.

> <b>Note</b>:  we consider currunt directory as a root. and all paths are given here is releted to the root directory. 

### a. Postgres setup
- If you don't have a docker please install it.
- Second we have docker-compose.storage.yml file for database. 

```
   $ root > docker compose -f docker-compose.storage.yml up -d
```

### b. Backend setup
- the backend server is written in golang and we choose the framework fiber for it. 
- we choose the ORM goqu here. 
- go uses the binary to run, and in development it's not possible to rerun on change so we will use air for a filewatcher, it will help us in development mode, so we should not each time build and run the binary by our self.

#### i. golang and fiber
- you need to download golang for you local system.
- then go to server directory and download the dependancies.
```
$ root/server > go get
```

- this will download the fiber and other required package.



#### ii. air (watcher)
- download the golang air for watcher. 
- we have air.tomb file for the configuration. 
```
$ root > go install github.com/air-verse/air@latest
```

- then you need to add golang binary to your path.
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

- start an watcher
```
$ root/server > air
```

#### iii. migration (before server run)
- we managed the make file for the commands which usavally need in production.
- currunt migration setup is not sure, it will take this way in future or change. but for now below utilities are available. 

1. make mfile to generate new migration file
2. make mup to migration up 
3. make mdown to migration down.

```
$ root/server > make mfile create_category_table
$ root/server > make mup
$ root/server > make mdown
```

> note: it's always recommended that before starting the server run migration up.
> note: it should run per deployment, not per server.
> note: make sure envs are properly configured.

#### iv. server (app)
- to start dev server, you need to do.
```
$ root/server > make dev
```

### c. Frontend setup

#### i. node and vue3

- we use node environment. here so please install the node.
- we use pnpm as a package manager so install that as well. 
> npx get-pnpm

- app directory have the code for frontend.
- first of all you need to install the dependacyes of the project.
```
$ root/app > pnpm install 
```

#### ii. server (watcher)

- you are ready to run this project. for that run below command. 
```
$ root/app > pnpm dev 
```

