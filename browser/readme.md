# Snapper

This is the javascript support library for the [snapper](https://github.com/tangerilli/snapper) project. It supports interacting with the snapper web-service to generate a PDF of a webpage, either by providing the URL (for publicly accessible pages) or by providing an HTML blob (for pages which require authentication).

## Usage

### Basic

```
<script type="text/javascript" src="path/to/snapper.min.js"></script>
<button onclick="snapper.convertPageToPdf({save: {filename: 'example.pdf'}});">Print PDF</button>
```

The above will initiate the download of a PDF containing the contents of the current page.

### Advanced

## Methods

## Options