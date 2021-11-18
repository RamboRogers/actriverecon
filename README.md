# ActiveRecon

ActiveRecon is a tool for maintaining a constant view of your network and systems/services.  It is different from other presentations layers in the primary way that it processes scan data into a database.  This allows multiple scans to merge together in a seamless way.  This allow for work containization and failures in scanning to be worked through without scrapping good scan data.

1. The goal is to deliver a clear, concise, simplistic view of a complicated landscape. This will be delivered in a web browser.
2. The output must is readable by anyone with basic network/application understanding.
4. New devices/old devices will be tracked.  Changes will be merged gracefully.
5. Screenshots are generated for HTTP/HTTPS based services.

## Usage
1. Run the application on a
2. READ output, sort, load into database using logic.
3. Render using logic in HTML.

## Build
1. Copy the sourcecode to a folder
2. 

## Packages
* https://gorm.io
* https://github.com/labstack/echo 
* https://gorm.io/driver/sqlite
* https://github.com/sensepost/gowitness
* https://github.com/robertdavidgraham/masscan

## Installation

Copy the code, compile, execute.

```bash
git clone

go build -o activerecon main.go

./activerecon
```
## Use Docker
Document the docker usage here...
```bash
docker build --no-cache -t activerecon .

docker run --name activerecon --restart always -d -p 9009:9009 activerecon
```

## Usage
Screenshots and examples here...

## Buy me a coffee
If you feel so inclined as to support my projects. Here's your chance! Thanks 
<a href="https://www.buymeacoffee.com/matthewrogers" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>
- [matt@matthewrogers.org]

## License
[GPLv3](https://choosealicense.com/licenses/agpl-3.0/)