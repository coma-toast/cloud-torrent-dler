var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
function getActiveTabURL() {
    return __awaiter(this, void 0, void 0, function* () {
        const tabs = yield chrome.tabs.query({
            currentWindow: true,
            active: true
        });
        return tabs[0];
    });
}
// sendMessage example
// const onPlay = async (e) => {
//     const bookmarkTime =
//         e.target.parentNode.parentNode.getAttribute("timestamp");
//     const activeTab = await getActiveTabURL();
//     chrome.tabs.sendMessage(activeTab.id, {
//         type: "PLAY",
//         value: bookmarkTime
//     });
// };
// * HERE
const refresh = document.getElementById("refresh");
refresh.addEventListener("click", () => __awaiter(this, void 0, void 0, function* () {
    console.log("Refreshing");
    alert("Refresh");
    options;
    // chrome.tabs.sendMessage((tabId, tab));
}));
document.addEventListener("DOMContentLoaded", () => __awaiter(this, void 0, void 0, function* () {
    console.log("Dom loaded");
    const activeTab = yield getActiveTabURL();
    if (activeTab.url.includes("magnetdl.com")) {
        const refresh = document.getElementById("refresh");
        refresh.hidden = false;
    }
    else {
        const container = document.getElementById("main-text");
        container.innerHTML =
            '<div id="main-text" class="font-medium">This is not a MagnetDL page.</div>';
    }
}));
