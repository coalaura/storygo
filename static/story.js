(() => {
	const $page = document.getElementById("page"),
		$context = document.getElementById("context"),
		$preview = document.getElementById("preview"),
		$image = document.getElementById("image"),
		$export = document.getElementById("export"),
		$markdown = document.getElementById("markdown"),
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

	function setImage(hash) {
		image = hash;

		if (image) {
			$preview.classList.add("image");
			$preview.style.backgroundImage = `url("/image/${image}")`;
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

			$context.setAttribute("disabled", "true");
			$text.setAttribute("disabled", "true");
			$direction.setAttribute("disabled", "true");
		} else {
			$page.classList.remove("generating");

			$context.removeAttribute("disabled");
			$text.removeAttribute("disabled");
			$direction.removeAttribute("disabled");
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

	$generate.addEventListener("click", async () => {
		if (uploading) return;

		if (generating) {
			controller?.abort();

			return;
		}

		const payload = {
			context: $context.value.trim(),
			text: $text.value.trim(),
			direction: $direction.value.trim(),
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

		store("context", payload.context);
		store("text", payload.text);
		store("direction", payload.direction);

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

			if (payload.text) {
				$text.value += "\n\n";
			}

			while (true) {
				const { value, done } = await reader.read();

				if (done) break;

				const chunk = decoder.decode(value, {
					stream: true,
				});

				$text.value = ($text.value + chunk).replace(/ {2,}/g, " ");
				$text.scrollTop = $text.scrollHeight;

				store("text", $text.value);
			}
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(`${err}`);
			}
		} finally {
			setGenerating(false);
		}
	});

	$markdown.addEventListener("click", () => {
		if (generating) return;

		const text = $text.value.trim();

		if (!text) return;

		download("story.md", "text/markdown", text);
	});

	$export.addEventListener("click", () => {
		if (generating) return;

		download(
			"story.json",
			"application/json",
			JSON.stringify(
				{
					context: $context.value.trim(),
					story: $text.value.trim(),
					direction: $direction.value.trim(),
				},
				null,
				4,
			),
		);
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
		store("context", $context.value.trim());
	});

	$text.addEventListener("change", () => {
		store("text", $text.value.trim());
	});

	$direction.addEventListener("change", () => {
		store("direction", $direction.value.trim());
	});

	$context.value = load("context", "");
	$text.value = load("text", "");
	$direction.value = load("direction", "");

	setImage(load("image", null));
})();
