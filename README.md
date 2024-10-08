# LibGen API

This repository contains a Go script that allows you to search for and download ebooks from Library Genesis.

## Installation

1. **Clone the repository:**

   
   ```bash
   git clone https://github.com/alvinmurimi/libgen.git
   cd libgen
   ```
   

2. **Install dependencies:**

   Ensure you have Go installed on your machine. You can download it [here](https://go.dev/dl/).

   
   ```bash
   go mod init github.com/alvinmurimi/libgen
   go mod tidy
   ```
   

## Usage

### Running the script

To run the script, use the following command:


```bash
go run main.go
```


This will start a server at http://localhost:8080.

## Running as a microservice
To run the application as a microservice, build and deploy it to your preferred cloud or container platform. Below are basic steps using Docker:

1. **Build the image:**
```bash
docker build -t libgen .
```

2. **Run the image:**
```bash
docker run -p 8080:8080 libgen
```
Replace `-p 8080:8080` with your preferred port mapping.


### Endpoints

The script exposes the following endpoints:

1. `GET /`
    - **Description**: Basic health check endpoint.
    - **Response**: Hello, World!

2. `GET /search`
    - **Description**: Search for eBooks by name.
    - **Query Parameters**:
      - **ebook**: The name of the eBook to search for.
      - **page**: (Optional) The page number of search results to retrieve. Defaults to 1.
    - **Response**: A JSON array of books with details such as author, title, URL, pages, size, language, category, and extension.

Example:


```bash
curl 'http://127.0.0.1:8080/search?ebook=john%20kiriamiti'
```
Output:
```json
[
  {
    "authors": [
      "John Kiriamiti"
    ],
    "title": "My Life with a Criminal: Milly's Story",
    "url": "http://library.lol/main/385AB0FBDD37033748A9E26F5BFD2D1F",
    "publisher": "Nairobi : Spear Books",
    "year": "1989",
    "pages": "",
    "size": "361 Kb",
    "language": "English",
    "category": "main",
    "extension": "epub"
  }
]
```


3. `GET /download`

    - **Description**: Retrieve download information for a specific eBook.
    - **Query Parameters**:
      - `ebook`: The URL of the eBook to get download information for.
    - **Response**: A JSON object with download links and details such as description, title, author, cloudflare link, IPFSIO link, and thumbnail.


Example:
```bash
curl 'http://127.0.0.1:8080/download?ebook=https://library.lol/main/385AB0FBDD37033748A9E26F5BFD2D1F'
```
Output:
```json
{
  "description": "John Kiriamiti's best-selling novel My Life in Crime has become a classic. Here Milly, his girlfriend, tells the poignant story of her life with the bank robber. They were in love, and he was gentle, kind and considerate. But after she moved in with him, she discovered his double life. She remained devoted, but the stress of his life bore its toll, and finally they parted. This sequel novel is also a bestseller in Kenya",
  "title": "My Life with a Criminal: Milly's Story",
  "url": "https://download.library.lol/main/3532000/385ab0fbdd37033748a9e26f5bfd2d1f/John%20Kiriamiti%20-%20My%20Life%20with%20a%20Criminal_%20Milly%27s%20Story-Nairobi%20_%20Spear%20Books%20%281989%29.epub",
  "author": "John Kiriamiti",
  "cloudflare": "https://cloudflare-ipfs.com/ipfs/bafykbzaced3hyqipvpkjei2d6cqy2qecjre77rusbuend2d2fvvdr5dch2phe?filename=John%20Kiriamiti%20-%20My%20Life%20with%20a%20Criminal_%20Milly%27s%20Story-Nairobi%20_%20Spear%20Books%20%281989%29.epub",
  "ipfsio": "https://gateway.ipfs.io/ipfs/bafykbzaced3hyqipvpkjei2d6cqy2qecjre77rusbuend2d2fvvdr5dch2phe?filename=John%20Kiriamiti%20-%20My%20Life%20with%20a%20Criminal_%20Milly%27s%20Story-Nairobi%20_%20Spear%20Books%20%281989%29.epub",
  "thumbnail": "http://library.lol/covers/3532000/385ab0fbdd37033748a9e26f5bfd2d1f-g.jpg"
}
```


## Dependencies

- [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery): HTML parsing and querying in Go.
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin): Web framework for Go.
- [github.com/parnurzeal/gorequest](https://github.com/parnurzeal/gorequest): Simplified HTTP client based on Go's net/http.

## Contributing
Contributions are welcome. Please open an issue or a pull request.