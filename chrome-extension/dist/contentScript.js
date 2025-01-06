var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
(() => {
    chrome.runtime.onMessage.addListener((obj, sender, response) => {
        const { type, value, server } = obj;
        if (type === "VALID" || type === "REFRESH") {
            refreshData(server);
        }
    });
    //     debugButton.addEventListener("click", async () => {
    //         console.log("click");
    //         let [tab] = await chrome.tabs.query({
    //             active: true,
    //             currentWindow: true
    //         });
    //         chrome.scripting.executeScript({
    //             target: { tabId: tab.id },
    //             func: initStorageCache
    //         });
    //     });
    //     function initStorageCache() {
    //         console.log("Refreshing cache");
    //         chrome.storage.sync.get("options", (data) => {
    //             Object.assign(options, data.options);
    //             console.log(options.serverUrl);
    //         });
    //     }
    function refreshData(target) {
        console.log("Refresh data in contentScript.js");
        const magnetLinks = document.getElementsByClassName("m");
        Array.from(magnetLinks, (magnet) => {
            console.log(magnet);
            const link = magnet.querySelector("a").href;
            console.log(link);
            const button = makeButton(target, link);
            magnet.append(button);
        });
    }
    function makeButton(target, magnet) {
        const magnetData = { link: magnet };
        const icon = document.createElement("img");
        icon.src = chrome.runtime.getURL("/images/cloud-storage-16x16.png");
        icon.onclick = () => {
            icon.src = chrome.runtime.getURL("/images/loading.png");
            postData(target, magnetData).then((response) => {
                if (response.status == 200) {
                    icon.src = chrome.runtime.getURL("/images/cloud-storage-16x16.png");
                }
                else {
                    icon.src = chrome.runtime.getURL("/images/error.png");
                    console.log(response);
                }
            });
        };
        return icon;
    }
    function postData(target, data) {
        return __awaiter(this, void 0, void 0, function* () {
            const response = yield fetch(`${target}/api/magnet`, {
                method: "POST",
                body: JSON.stringify(data)
            });
            return response.json();
        });
    }
})();
