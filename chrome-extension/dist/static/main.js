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
    let downloadMagnetButtons = document.getElementsByClassName("download-magnet");
    Array.from(downloadMagnetButtons).forEach((button) => {
        button.addEventListener("click", function () {
            let formData = new FormData();
            formData.append("link", this.dataset.torrent);
            let data = JSON.stringify(Object.fromEntries(formData));
            button.classList.remove("btn-success");
            button.classList.add("btn-danger");
            let resolution = button.innerHTML;
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
    let searchButton = document.getElementById("search-button");
    if (!searchButton) {
        throw new Error("No search button found");
    }
    searchButton.addEventListener("click", function () {
        let formData = new FormData();
        const search = document.getElementById("search");
        if (!search) {
            throw new Error("No search input found");
        }
        formData.append("search", search.value);
        let data = JSON.stringify(Object.fromEntries(formData));
        apiRequest("GET", "search", data)
            .then((response) => {
            if (!response) {
                throw new Error("error searching");
            }
            return response.json();
        })
            .then((responseData) => {
            console.log(responseData);
        })
            .catch((err) => {
            console.error(err);
        });
    });
    // interface MovieData = {
    //     imdb: string, title: string, image: string, year: string,rating: string
    // }
    // function movieRow();
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
