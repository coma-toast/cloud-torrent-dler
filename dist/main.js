var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
document.addEventListener("DOMContentLoaded", function () {
    let navButtons = document.getElementsByClassName("nav-link");
    Array.from(navButtons).forEach((navButton) => {
        navButton.addEventListener("click", function () { });
    });
    let downloadButtons = document.getElementsByClassName("download");
    Array.from(downloadButtons).forEach((button) => {
        button.addEventListener("click", function () {
            let formData = new FormData();
            formData.append("link", this.dataset.torrent);
            formData.append("auto_download", "movie");
            let data = JSON.stringify(Object.fromEntries(formData));
            button.classList.remove("btn-success");
            button.classList.add("btn-danger");
            let resolution = button.innerHTML;
            button.innerHTML = `<span class="spinner-border spinner-border-sm"></span>`;
            apiRequest("POST", "torrent", data)
                .then((response) => {
                if (!response) {
                    throw new Error("error posting torrent");
                }
                return response.json();
            })
                .then((responseData) => {
                if (responseData) {
                    button.innerHTML = resolution;
                    button.classList.remove("btn-danger");
                    button.classList.add("btn-warning");
                }
                else {
                    button.innerHTML = "Failed";
                }
                console.log(responseData);
            })
                .catch((e) => console.error("Error posting magnet", e));
        });
    });
    let downloadMagnetButtons = document.getElementsByClassName("download-magnet");
    Array.from(downloadMagnetButtons).forEach((button) => {
        button.addEventListener("click", function () {
            const formData = new FormData();
            formData.append("link", this.dataset.torrent);
            formData.append("auto_download", "show");
            const data = JSON.stringify(Object.fromEntries(formData));
            button.classList.remove("btn-success");
            button.classList.add("btn-danger");
            const resolution = button.innerHTML;
            button.innerHTML = `<span class="spinner-border spinner-border-sm"></span>`;
            apiRequest("POST", "magnet", data)
                .then((response) => {
                if (!response) {
                    throw new Error("error posting magnet");
                }
                return response.json();
            })
                .then((responseData) => {
                if (responseData.result) {
                    button.innerHTML = resolution;
                    button.classList.remove("btn-danger");
                    button.classList.add("btn-warning");
                }
                else {
                    button.innerHTML = "Failed";
                }
                console.log(responseData);
            })
                .catch((e) => console.error("Error posting magnet", e));
        });
    });
    const searchButton = document.getElementById("search-button");
    if (!searchButton) {
        throw new Error("No search button found");
    }
    searchButton.addEventListener("click", function (event) {
        event.preventDefault();
        const message = document.getElementById("search-message");
        const tableBody = document.getElementById("movie-table-body");
        if (!tableBody) {
            throw new Error("No table found");
        }
        console.log(tableBody);
        while (tableBody.firstChild) {
            tableBody.removeChild(tableBody.firstChild);
        }
        const formData = new FormData();
        const search = document.getElementById("search");
        if (!search) {
            throw new Error("No search input found");
        }
        formData.append("search", search.value);
        // const data = JSON.stringify(Object.fromEntries(formData));
        apiRequest("GET", "search", formData)
            .then((response) => {
            if (!response) {
                message.innerHTML = "Error searching";
                throw new Error("error searching");
            }
            return response.json();
        })
            .then((responseData) => {
            console.log(responseData);
            if (responseData.movie_count === 0) {
                message.innerHTML = "No movies found";
                return;
            }
            responseData.movies.map((movie) => {
                tableBody.appendChild(movieRow({
                    imdb: movie.imdb_code,
                    title: movie.title,
                    image: movie.medium_cover_image,
                    year: movie.year.toString(),
                    rating: movie.rating.toString(),
                    torrents: movie.torrents
                }));
            });
        })
            .catch((err) => {
            console.error(err);
        });
    });
    function movieRow(data) {
        const row = document.createElement("tr");
        row.appendChild(imageCell(data.image, data.title, data.imdb));
        row.appendChild(textCell(data.year));
        row.appendChild(textCell(data.rating));
        row.appendChild(infoCell(data.torrents));
        return row;
    }
    function imageCell(image, title, imdb) {
        const cell = document.createElement("td");
        const p = document.createElement("p");
        const titleBold = document.createElement("b");
        titleBold.innerHTML = title;
        p.appendChild(titleBold);
        const img = document.createElement("img");
        img.src = image;
        const a = document.createElement("a");
        a.href = `https://www.imdb.com/title/${imdb}`;
        a.appendChild(img);
        cell.appendChild(p);
        cell.appendChild(a);
        return cell;
    }
    function textCell(text) {
        const cell = document.createElement("td");
        cell.innerHTML = text;
        return cell;
    }
    function infoCell(torrents) {
        const cell = document.createElement("td");
        const table = document.createElement("table");
        table.classList.add("table", "table-striped");
        const thead = document.createElement("thead");
        const theadRow = document.createElement("tr");
        const qualityHeader = document.createElement("th");
        qualityHeader.innerHTML = "Quality";
        const seedsHeader = document.createElement("th");
        seedsHeader.innerHTML = "Seeds";
        const peersHeader = document.createElement("th");
        peersHeader.innerHTML = "Peers";
        const dateUploadedHeader = document.createElement("th");
        dateUploadedHeader.innerHTML = "Date Uploaded";
        theadRow.appendChild(qualityHeader);
        theadRow.appendChild(seedsHeader);
        theadRow.appendChild(peersHeader);
        theadRow.appendChild(dateUploadedHeader);
        thead.appendChild(theadRow);
        table.appendChild(thead);
        const tbody = document.createElement("tbody");
        torrents.forEach((torrent) => {
            tbody.appendChild(torrentRow(torrent));
        });
        table.appendChild(tbody);
        cell.appendChild(table);
        return cell;
    }
    function torrentRow(torrent) {
        const row = document.createElement("tr");
        const urlCell = document.createElement("td");
        const qualityCell = document.createElement("td");
        const seedsCell = document.createElement("td");
        const peersCell = document.createElement("td");
        const dateUploadedCell = document.createElement("td");
        const downloadCell = document.createElement("td");
        const downloadButton = document.createElement("button");
        downloadButton.classList.add("btn", "btn-success", "download");
        downloadButton.dataset.torrent = torrent.url;
        downloadButton.innerHTML = "Download";
        downloadCell.appendChild(downloadButton);
        urlCell.innerHTML = torrent.url;
        qualityCell.innerHTML = torrent.quality;
        seedsCell.innerHTML = torrent.seeds.toString();
        peersCell.innerHTML = torrent.peers.toString();
        dateUploadedCell.innerHTML = torrent.date_uploaded;
        row.appendChild(urlCell);
        row.appendChild(qualityCell);
        row.appendChild(seedsCell);
        row.appendChild(peersCell);
        row.appendChild(dateUploadedCell);
        row.appendChild(downloadCell);
        return row;
    }
    /**
     * * This is the main API call function
     *
     * @param   {string}                method  "GET", "POST"
     * @param   {string}                target
     * @param   {FormData}              data  new FormData
     * @param   {Object|string[[]]}     params   {clientID: 1} | [['clientID', '1']]
     *
     * @return  {Response}
     */
    function apiRequest(method, target, data) {
        return __awaiter(this, void 0, void 0, function* () {
            try {
                let response = new Response();
                let headers = new Headers();
                /*
                 * Fetch allows relative URLs, but you can't have a body for GET requests.
                 * So we have to build full URL with query params for GET requests with no body (not even `null`)
                 * and then use the normal fetch request for POSTs
                 */
                if (method === "GET") {
                    const params = new URLSearchParams(data).toString();
                    const url = new URL(`${window.location.origin}/api/${target}?${params}`);
                    response = yield fetch(url, {
                        method,
                        mode: "cors",
                        headers
                    });
                }
                else {
                    console.log(data);
                    response = yield fetch("/api/" + target, {
                        method,
                        mode: "cors",
                        headers,
                        body: data
                    });
                }
                console.debug(response);
                return response;
            }
            catch (error) {
                console.error(error);
            }
        });
    }
});
