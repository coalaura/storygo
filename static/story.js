(() => {
	const $page = document.getElementById("page"),
		$context = document.getElementById("context"),
		$preview = document.getElementById("preview"),
		$image = document.getElementById("image"),
		$import = document.getElementById("import"),
		$importFile = document.getElementById("import-file"),
		$export = document.getElementById("export"),
		$markdown = document.getElementById("markdown"),
		$delete = document.getElementById("delete"),
		$text = document.getElementById("text"),
		$direction = document.getElementById("direction"),
		$generate = document.getElementById("generate");

	let uploading, generating, image;

	function setUploading(status) {
		uploading = status;

		if (uploading) {
			$preview.classList.add("uploading");
			$generate.setAttribute("disabled", "true");
		} else {
			$preview.classList.remove("uploading");
			$generate.removeAttribute("disabled");
		}
	}

	let exists;

	function setImage(hash) {
		exists?.abort?.();

		image = hash;

		if (image) {
			$preview.classList.add("image");
			$preview.style.backgroundImage = `url("/image/${image}")`;

			exists = new AbortController();

			fetch(`/image/${image}`, {
				signal: exists.signal,
			})
				.then((resp) => {
					if (!resp.ok) {
						setImage(null);
					}
				})
				.catch((err) => {
					if (err.name === "AbortError") {
						return;
					}

					setImage(null);
				});
		} else {
			$preview.classList.remove("image");
			$preview.style.backgroundImage = "";
		}

		store("image", hash);
	}

	function setGenerating(status) {
		generating = status;

		if (generating) {
			$page.classList.add("generating");

			$text.setAttribute("disabled", "true");
		} else {
			$page.classList.remove("generating");

			$text.removeAttribute("disabled");
		}
	}

	function download(name, type, data) {
		const blob = new Blob([data], {
			type: type,
		});

		const url = URL.createObjectURL(blob),
			a = document.createElement("a");

		a.style.display = "none";
		a.href = url;
		a.download = name;

		document.body.appendChild(a);

		a.click();

		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	function clean(text) {
		text = text.trim();
		text = text.replace(/ {2,}/g, " ");

		return text;
	}

	function storeAll() {
		store("context", clean($context.value));
		store("text", clean($text.value));
		store("direction", clean($direction.value));
	}

	function store(name, value) {
		if (!value) {
			localStorage.removeItem(name);

			return;
		}

		localStorage.setItem(name, value);
	}

	function load(name, def) {
		const value = localStorage.getItem(name);

		if (!value) {
			return def;
		}

		return value;
	}

	let controller;

	async function generate(inline) {
		if (uploading) return;

		if (generating) {
			controller?.abort?.();

			return;
		}

		const payload = {
			context: clean($context.value),
			text: clean($text.value) + (inline ? " " : "\n\n"),
			direction: clean($direction.value),
			image: image,
		};

		if (!payload.context) {
			alert("Missing context.");

			return;
		}

		setGenerating(true);

		controller = new AbortController();

		$context.value = payload.context;
		$text.value = payload.text;
		$direction.value = payload.direction;

		storeAll();

		try {
			const response = await fetch("/generate", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
				signal: controller.signal,
			});

			if (!response.ok) {
				throw new Error(`Generation failed with status ${response.status}`);
			}

			const reader = response.body.getReader(),
				decoder = new TextDecoder();

			while (true) {
				const { value, done } = await reader.read();

				if (done) break;

				const chunk = decoder.decode(value, {
					stream: true,
				});

				payload.text += chunk;

				$text.value = payload.text;
				$text.scrollTop = $text.scrollHeight;

				store("text", payload.text);
			}
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(`${err}`);
			}
		} finally {
			setGenerating(false);

			payload.text = clean(dai.clean(payload.text));

			$text.value = payload.text;

			store("text", payload.text);
		}
	}

	$generate.addEventListener("click", async () => {
		generate(false);
	});

	$markdown.addEventListener("click", () => {
		if (generating) return;

		const text = clean($text.value);

		if (!text) return;

		download("story.md", "text/markdown", text);
	});

	$export.addEventListener("click", () => {
		if (generating) return;

		download(
			"save-file.json",
			"application/json",
			JSON.stringify({
				ctx: clean($context.value),
				txt: clean($text.value),
				dir: clean($direction.value),
				img: image,
			}),
		);
	});

	$import.addEventListener("click", () => {
		if (generating || uploading) return;

		$importFile.click();
	});

	$importFile.addEventListener("change", (event) => {
		const file = event.target.files[0];

		if (!file) return;

		const reader = new FileReader();

		reader.onload = (e) => {
			try {
				const json = JSON.parse(e.target.result);

				if (!json?.ctx && !json?.txt && !json?.dir) {
					throw new Error("empty safe file");
				}

				if (json.ctx && typeof json.ctx === "string") {
					$context.value = clean(json.ctx);
				} else {
					$context.value = "";
				}

				if (json.txt && typeof json.txt === "string") {
					$text.value = clean(json.txt);
					$text.scrollTop = $text.scrollHeight;
				} else {
					$text.value = "";
				}

				if (json.dir && typeof json.dir === "string") {
					$direction.value = clean(json.dir);
				} else {
					$direction.value = "";
				}

				if (json.img && typeof json.img === "string") {
					setImage(json.img);
				} else {
					setImage(null);
				}

				storeAll();
			} catch {
				alert("Invalid safe file.");
			}
		};

		reader.readAsText(file);
	});

	$delete.addEventListener("click", () => {
		if (uploading || generating) return;

		if (!confirm("Are you sure you want to clear the story, context, directions and image?")) return;

		$context.value = "";
		$text.value = "";
		$direction.value = "";

		setImage(null);

		storeAll();
	});

	$preview.addEventListener("click", () => {
		if (uploading || generating) return;

		setImage(null);

		$image.click();
	});

	$image.addEventListener("change", async (event) => {
		const file = event.target.files[0];

		if (!file) {
			return;
		}

		setUploading(true);

		const form = new FormData();

		form.append("details", prompt("Important details in the image (optional)", "") || "");
		form.append("image", file);

		let hash;

		try {
			const response = await fetch("/upload", {
				method: "POST",
				body: form,
			});

			if (!response.ok) {
				throw new Error(`Upload failed with status ${response.status}`);
			}

			hash = await response.text();
		} catch (err) {
			alert(`${err}`);

			return;
		} finally {
			setUploading(false);
		}

		setImage(hash);
	});

	$context.addEventListener("change", () => {
		store("context", clean($context.value));
	});

	$context.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "Enter") {
			event.preventDefault();

			$generate.click();
		}
	});

	$text.addEventListener("change", () => {
		store("text", clean($text.value));
	});

	$text.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "Enter") {
			event.preventDefault();

			$generate.click();
		} else if (event.key === "Tab") {
			event.preventDefault();

			generate(true);
		}
	});

	$direction.addEventListener("change", () => {
		store("direction", clean($direction.value));
	});

	$direction.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "Enter") {
			event.preventDefault();

			$generate.click();
		}
	});

	document.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "s") {
			event.preventDefault();

			$markdown.click();
		}
	});

	$context.value = load("context", "");
	$text.value = load("text", "");
	$direction.value = load("direction", "");

	$text.scrollTop = $text.scrollHeight;

	setImage(load("image", null));
})();
