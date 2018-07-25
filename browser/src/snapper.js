import {version} from '../package.json';

let defaults = {
    printServiceURL: 'http://localhost:8088/'
};

function decodeAndSaveBase64Pdf(base64Data, filename, successCallback, errorCallback) {
    // Create a new Blob object using the base64-encoded data
    const byteCharacters = atob(base64Data);
    const byteNumbers = new Array(byteCharacters.length);

    for (let i = 0; i < byteCharacters.length; i++) {
        byteNumbers[i] = byteCharacters.charCodeAt(i);
    }

    const byteArray = new Uint8Array(byteNumbers);
    const blob = new Blob([byteArray], {type: 'image/pdf'});

    // Create a link element, hide it, direct
    // it towards the blob, and then 'click' it programatically
    const a = document.createElement('a');

    a.style = 'display: none';
    document.body.appendChild(a);
    // Create a DOMString representing the blob
    // and point the link element towards it
    const url = window.URL.createObjectURL(blob);

    a.href = url;
    a.download = filename;
    // programatically click the link to trigger the download
    a.click();
    // release the reference to the file by revoking the Object URL
    window.URL.revokeObjectURL(url);
    if (successCallback !== undefined) {
        successCallback();
    }
}

function inlineCanvases(doc, shadowDoc) {
    shadowDoc = shadowDoc || doc;

    // Replace any canvas elements with images because the canvas data won't be included in the HTML
    // and therefore won't be printed.
    const canvases = doc.getElementsByTagName('canvas');
    const shadowCanvases = shadowDoc.getElementsByTagName('canvas');

    if (canvases.length !== shadowCanvases.length) {
        console.error('Error trying to inline canvases, doc and shadowDoc differ');
    }
    // document.getElementsByTagName seems to return some sort of dynamic list, because as we
    // iterate over it and replace the canvas elements, the iteration doesn't work properly (only
    // half the items actually get replaced).  To fix that, we copy all the canvases to a simple list
    // and then iterate over that
    const stableShadowCanvasList = [...shadowCanvases];
    const stableCanvasList = [...canvases];

    stableCanvasList.forEach((canvas, index) => {
        const data = canvas.toDataURL();
        const img = shadowDoc.createElement('img');

        img.width = canvas.width;
        img.height = canvas.height;
        img.src = data;

        const shadowCanvas = stableShadowCanvasList[index];
        const parent = shadowCanvas.parentNode;

        parent.replaceChild(img, shadowCanvas);
    });
}

function removeScriptTags(doc) {
    const scripts = doc.getElementsByTagName('script');
    const stableScriptList = [...scripts];

    stableScriptList.forEach((script, index) => {
        const parent = script.parentNode;

        parent.removeChild(script);
    });
    return doc;
}

function inlineCssBlobs(doc) {
    // Replace blobs with inline CSS (this is mainly useful in a development environment
    // using webpack)
    let pendingReplacements = 0;
    let resolver = null;
    // TODO: Apparently you can't do this in IE?
    const blobPromise = new Promise((resolve, reject) => {resolver = resolve;});

    function replaceBlob(element) {
        const xhr = new XMLHttpRequest();

        xhr.open('GET', element.href, true);
        xhr.responseType = 'blob';
        xhr.onload = function (e) {
            if (this.status === 200) {
                const reader = new FileReader();

                reader.addEventListener('loadend', function () {
                    const decoder = new TextDecoder('utf-8');
                    const result = decoder.decode(reader.result);
                    const inlineCss = doc.createElement('style');

                    inlineCss.type = 'text/css';

                    if (inlineCss.styleSheet) {
                        inlineCss.styleSheet.cssText = result;
                    } else {
                        inlineCss.appendChild(doc.createTextNode(result));
                    }

                    const parent = element.parentNode;

                    parent.replaceChild(inlineCss, element);

                    if (--pendingReplacements === 0) {
                        resolver();
                    }
                });
                reader.readAsArrayBuffer(this.response);
            }
        };
        xhr.send();
    }

    const links = doc.getElementsByTagName('link');
    const stableLinks = [];

    for (let i = 0; i < links.length; i++) {
        stableLinks.push(links[i]);
        if (links[i].href.startsWith('blob')) {
            pendingReplacements++;
        }
    }

    for (let i = 0; i < stableLinks.length; i++) {
        if (links[i].href.startsWith('blob')) {
            replaceBlob(stableLinks[i]);
        }
    }
    if (pendingReplacements === 0) {
        resolver();
    }
    return blobPromise;
}

function getShadowDoc() {
    const parser = new DOMParser();

    return parser.parseFromString(document.documentElement.innerHTML, 'text/html');
}

function getDocumentWithInlinedCanvases() {
    const shadowDoc = getShadowDoc();

    inlineCanvases(document, shadowDoc);
    return shadowDoc;
}

function printToPdf(url, params, options) {
    const xhr = new XMLHttpRequest();

    xhr.onload = function (e) {
        if (this.status === 200) {
            const response = this.response;

            if (options.base64DataCallback !== undefined) {
                options.base64DataCallback(response.pdfData);
            } else {
                const saveOptions = options.save || {};
                const filename = saveOptions.filename || 'page.pdf';

                decodeAndSaveBase64Pdf(
                    response.pdfData,
                    filename,
                    saveOptions.successCallback,
                    options.errorHandler
                );
            }
        } else {
            if (options.errorHandler !== undefined) {
                options.errorHandler('Error generating PDF', e);
            }
        }
    };

    xhr.onerror = (e) => {
        if (options.errorHandler) {
            options.errorHandler('Error generating PDF', e);
        }
    };

    xhr.open('POST', url);
    xhr.responseType = 'json';
    xhr.setRequestHeader('content-type', 'application/json');
    xhr.send(params);
}

function convertUrlToPdf(url, options) {
    options = options || {};

    const data = JSON.stringify({
        'url': url,
        'printParameters': options.printParameters
    });

    printToPdf(options.printServiceURL || defaults.printServiceURL, data, options);
}

function convertHtmlToPdf(html, options) {
    options = options || {};

    const data = JSON.stringify({
        'html': html,
        'printParameters': options.printParameters
    });

    const xhr = new XMLHttpRequest();

    xhr.onload = function (e) {
        if (this.status === 200) {
            const response = this.response;

            if (options.base64DataCallback !== undefined) {
                options.base64DataCallback(response.pdfData);
            } else {
                const saveOptions = options.save || {};
                const filename = saveOptions.filename || 'page.pdf';

                decodeAndSaveBase64Pdf(
                    response.pdfData,
                    filename,
                    saveOptions.successCallback,
                    options.errorHandler
                );
            }
        } else {
            if (options.errorHandler !== undefined) {
                options.errorHandler('Error generating PDF', e);
            }
        }
    };

    xhr.onerror = (e) => {
        if (options.errorHandler) {
            options.errorHandler('Error generating PDF', e);
        }
    };

    const url = options.printServiceURL || defaults.printServiceURL;

    xhr.open('POST', url);
    xhr.responseType = 'json';
    xhr.setRequestHeader('content-type', 'application/json');
    xhr.send(data);
}

/*
Option format:
options = {
    errorHandler: function(errorMessage, errorInfo=null) {},
    printParameters: {
        ...TODO
    },
    base64DataCallback: function(pdfData),
    save: {
        filename: 'page.pdf',
        successCallback: function() {}
    },
    printServiceURL: 'http://localhost:8088/',
    inlineCanvas: true,
    inlineBlobs: false,
    removeScripts: false,
    preprocessor: function(html) { return html; }
}
*/
function convertPageToPdf(options) {
    options = options || {};
    // first inlines some stuff on the page, then turns the entire page into a PDF
    const shouldInlineCanvas = options.inlineCanvas !== false;
    const shouldInlineBlobs = options.inlineBlobs === true;
    const shouldRemoveScripts = options.removeScripts === true;
    const preprocessor = options.preprocessor || ((html) => { return html; });
    let shadowDoc = document;

    if (shouldInlineCanvas) {
        shadowDoc = getDocumentWithInlinedCanvases();
    }

    if (shouldRemoveScripts) {
        if (shadowDoc === document) {
            shadowDoc = getShadowDoc();
        }
        shadowDoc = removeScriptTags(shadowDoc);
    }

    if (shouldInlineBlobs) {
        const blobPromise = inlineCssBlobs(shadowDoc);

        blobPromise.then(() => {
            convertHtmlToPdf(preprocessor(shadowDoc.documentElement.innerHTML), options);
        }, () => {
            options.errorHandler('Error inlining blobs');
        });
    } else {
        convertHtmlToPdf(preprocessor(shadowDoc.documentElement.innerHTML), options);
    }
}

let snapper = {
    convertPageToPdf: convertPageToPdf,
    convertHtmlToPdf: convertHtmlToPdf,
    convertUrlToPdf: convertUrlToPdf,
    getDocumentWithInlinedCanvases: getDocumentWithInlinedCanvases,
    inlineCanvases: inlineCanvases,
    inlineCssBlobs: inlineCssBlobs,
    removeScriptTags: removeScriptTags,
    version: version,
    defaults: defaults
};

export default snapper;
