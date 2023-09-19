console.log("loaded options");

const saveButton = document.getElementById("save-server");
const target = document.getElementById("target") as HTMLInputElement;

// In-page cache of the user's options
const optionsData: Options = {
    serverUrl: "test"
};

// Initialize the form with the user's option settings
chrome.storage.sync.get("options", (data) => {
    console.log("getting data from options.js:", data);
    Object.assign(optionsData, data.options);
    console.log("target :>> ", target);
    target.value = optionsData.serverUrl;
});

saveButton.addEventListener("click", () => {
    let newOptions = {
        serverUrl: target.value
    };
    chrome.storage.sync.set({ options: newOptions }, () => {
        console.log("Saving URL", target.value);
    });
    chrome.storage.sync.get("options", (data) => {
        Object.assign(optionsData, data.options);
    });
});
