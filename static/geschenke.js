new Vue({
  el: "#vue",
  computed: {
    id() {
      return document.querySelector("input[name='id']").value;
    },
    maxSize() {
      return parseInt(document.querySelector("input[name='maxSize']").value);
    },
  },
  methods: {
    onBoxClick() {
      switch (this.state) {
        case "noFileSelected":
        case "fileSelected":
        case "error":
          this.$refs.fileInput.click();
          break;
        default:
      }
    },
    onBoxDragover(event) {
      switch (this.state) {
        case "noFileSelected":
        case "fileSelected":
        case "error":
          event.preventDefault();
          break;
        default:
      }
    },
    onBoxDrop(event) {
      event.preventDefault();
      this.readFile(event.dataTransfer.files[0]);
    },
    onFileInputChange(event) {
      this.readFile(this.$refs.fileInput.files[0]);
    },
    async onUploadClick(event) {
      event.stopPropagation();
      this.state = "busy";
      let response = await fetch(`/id/${this.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(this.file),
      });
      if (!response.ok) {
        this.error = "Failed to upload file.";
        this.state = "error";
        return;
      }
      this.state = "success";
    },
    readFile(file) {
      if (file.size > this.maxSize) {
        this.error = `File size (${(file.size / 1024 / 1024).toFixed(
          2
        )} MiB) exceeds maximum.`;
        this.state = "error";
        return;
      }

      let reader = new FileReader();
      reader.addEventListener("load", (event) => {
        this.file = {
          id: this.id,
          name: file.name,
          size: file.size,
          dataURL: event.target.result,
        };
        this.state = "fileSelected";
      });
      reader.readAsDataURL(file);
    },
  },
  data() {
    return {
      state: "noFileSelected",
      file: undefined,
      error: undefined,
    };
  },
});
