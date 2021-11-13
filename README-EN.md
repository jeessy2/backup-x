# backup-x
  A database backup tool with web interfaces.
  - [x] Support custom commands.
  - [x] Obsolete files will be deleted automatically.
  - [x] Support the backup files copy to simple data storage(s3).
  - [x] Automatic backup in everyday night.
  - [x] Webhook support

## use in docker
  ```
    docker run -d \
    --name backup-x \
    --restart=always \
    -p 9977:9977 \
    -v /opt/backup-x-files:/app/backup-x-files \
    jeessy/backup-x
  ```

  ![avatar](https://raw.githubusercontent.com/jeessy2/backup-x/master/backup-x-web.png)

