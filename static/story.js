(() => {
	const $page = document.getElementById("page"),
		$model = document.getElementById("model"),
		$context = document.getElementById("context"),
		$preview = document.getElementById("preview"),
		$image = document.getElementById("image"),
		$tag = document.getElementById("tag"),
		$tags = document.getElementById("tags"),
		$import = document.getElementById("import"),
		$importFile = document.getElementById("import-file"),
		$export = document.getElementById("export"),
		$markdown = document.getElementById("markdown"),
		$delete = document.getElementById("delete"),
		$text = document.getElementById("text"),
		$status = document.getElementById("status"),
		$mode = document.getElementById("mode"),
		$modeName = document.getElementById("mode-name"),
		$direction = document.getElementById("direction"),
		$suggest = document.getElementById("suggest"),
		$generate = document.getElementById("generate");

	let uploading,
		generating,
		suggesting,
		image,
		model,
		mode = "generate";

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

	function setStatus(status) {
		$status.textContent = status || "";
	}

	function setSuggesting(status) {
		suggesting = status;

		if (suggesting) {
			$direction.setAttribute("disabled", "true");
		} else {
			$direction.removeAttribute("disabled");
		}
	}

	function setMode(state) {
		mode = state || "generate";

		if (mode === "generate") {
			$modeName.textContent = "Story";
			$mode.title = "Switch to story outline/overview mode";
			$mode.classList.remove("overview");
		} else {
			$modeName.textContent = "Overview";
			$mode.title = "Switch to detailed story writing mode";
			$mode.classList.add("overview");
		}

		store("mode", mode);
	}

	function appendTag(content) {
		const tag = document.createElement("div");

		tag.textContent = content;
		tag.classList.add("tag");

		tag.addEventListener("click", () => {
			tag.remove();

			store("tags", buildTags());
		});

		$tags.appendChild(tag);
	}

	const modelList = [];

	function setModel(set) {
		if (!modelList.find((entry) => entry.key === set)) {
			set = modelList[0].key;
		}

		model = set;

		for (const entry of modelList) {
			const { key, element } = entry;

			if (key === model) {
				element.classList.add("selected");
			} else {
				element.classList.remove("selected");
			}
		}

		store("model", model);
	}

	function buildModels(models) {
		for (const entry of models) {
			const element = document.createElement("div");

			element.classList.add("model");

			const name = document.createElement("div");

			name.textContent = entry.name;
			name.classList.add("name");

			if (entry.vision) {
				name.classList.add("vision");
			}

			element.appendChild(name);

			const tags = document.createElement("div");

			tags.classList.add("tags");

			for (const tag of entry.tags) {
				const el = document.createElement("div");

				el.title = tag;
				el.style.backgroundImage = `url("/icons/tags/${tag}.svg")`;
				el.classList.add("tag");

				tags.appendChild(el);
			}

			element.appendChild(tags);

			$model.appendChild(element);

			element.addEventListener("click", () => {
				setModel(entry.key);
			});

			modelList.push({
				key: entry.key,
				element: element,
			});
		}

		setModel(load("model", null));
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
		text = text.trim().replace(/\r\n/g, "\n");

		text = text.replace(/ {2,}/g, " ");
		text = text.replace(/ +\n/g, "\n");

		return text;
	}

	function storeAll() {
		store("context", clean($context.value));
		store("text", clean($text.value));
		store("direction", clean($direction.value));
		store("tags", buildTags());
	}

	function store(name, value) {
		if (!value) {
			localStorage.removeItem(name);

			return;
		}

		if (typeof value === "object") {
			value = JSON.stringify(value);
		}

		localStorage.setItem(name, value);
	}

	function load(name, def) {
		const value = localStorage.getItem(name);

		if (!value) {
			return def;
		}

		if (value?.match?.(/^[{[]/m)) {
			try {
				return JSON.parse(value);
			} catch {}

			return null;
		}

		return value;
	}

	async function stream(url, options, callback, finished) {
		setStatus("Waiting...");

		try {
			const response = await fetch(url, options);

			if (!response.ok) {
				throw new Error(`Stream failed with status ${response.status}`);
			}

			const reader = response.body.getReader(),
				decoder = new TextDecoder();

			while (true) {
				const { value, done } = await reader.read();

				if (done) break;

				if (value.length === 1 && value[0] === 0) {
					setStatus("Reasoning...");

					continue;
				}

				setStatus("Writing...");

				const chunk = decoder.decode(value, {
					stream: true,
				});

				callback(chunk);
			}
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(`${err}`);
			}
		} finally {
			setStatus(false);

			finished();
		}
	}

	function buildTags() {
		const tags = [];

		$tags.querySelectorAll(".tag").forEach((tag) => {
			tags.push(tag.textContent.trim());
		});

		return tags;
	}

	function buildPayload(inline) {
		const payload = {
			model: model,
			context: clean($context.value),
			text: clean($text.value),
			direction: clean($direction.value),
			tags: buildTags(),
			image: image,
		};

		if (!payload.context) {
			alert("Missing context.");

			return false;
		}

		if (payload.text) {
			payload.text += inline ? " " : "\n\n";
		}

		$context.value = payload.context;
		$text.value = payload.text;
		$direction.value = payload.direction;

		$text.scrollTop = $text.scrollHeight;

		return payload;
	}

	let controller;

	async function generate(inline) {
		if (uploading || suggesting) return;

		if (generating) {
			controller?.abort?.();

			return;
		}

		if (mode === "overview") {
			$text.value = "";
		}

		const payload = buildPayload(inline);

		if (!payload) {
			return;
		}

		setGenerating(true);

		controller = new AbortController();

		storeAll();

		stream(
			`/${mode}`,
			{
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
				signal: controller.signal,
			},
			(chunk) => {
				payload.text += chunk;

				$text.value = payload.text;
				$text.scrollTop = $text.scrollHeight;

				store("text", payload.text);
			},
			() => {
				setGenerating(false);

				payload.text = clean(dai.clean(payload.text));

				$text.value = payload.text;

				store("text", payload.text);
			},
		);
	}

	$generate.addEventListener("click", async () => {
		generate(false);
	});

	$mode.addEventListener("click", () => {
		if (generating || suggesting) return;

		if (
			$text.value.trim() &&
			!confirm(
				"Are you sure you want to switch modes? This will clear the story field.",
			)
		)
			return;

		$text.value = "";

		if (mode === "overview") {
			setMode("generate");
		} else {
			setMode("overview");
		}
	});

	$markdown.addEventListener("click", () => {
		if (generating) return;

		const text = clean($text.value);

		if (!text) return;

		download("story.md", "text/markdown", text);
	});

	$export.addEventListener("click", () => {
		if (generating || suggesting) return;

		download(
			"save-file.json",
			"application/json",
			JSON.stringify({
				mdl: model,
				ctx: clean($context.value),
				txt: clean($text.value),
				dir: clean($direction.value),
				tgs: buildTags(),
				img: image,
				mde: mode || "generate",
			}),
		);
	});

	$import.addEventListener("click", () => {
		if (generating || uploading || suggesting) return;

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

				if (json.mdl && typeof json.mdl === "string") {
					setModel(json.mdl);
				} else {
					setModel(null);
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

				if (json.tgs && Array.isArray(json.tgs)) {
					for (const tag of json.tgs) {
						appendTag(tag);
					}
				} else {
					$tags.innerHTML = "";
				}

				if (json.img && typeof json.img === "string") {
					setImage(json.img);
				} else {
					setImage(null);
				}

				if (json.mde && typeof json.mde === "string") {
					setMode(json.mde);
				} else {
					setMode("generate");
				}

				storeAll();
			} catch (err) {
				console.error(err);

				alert("Invalid safe file.");
			}
		};

		reader.readAsText(file);
	});

	$delete.addEventListener("click", () => {
		if (uploading || generating || suggesting) return;

		if (
			!confirm(
				"Are you sure you want to clear the story, context, directions and image?",
			)
		)
			return;

		$context.value = "";
		$text.value = "";
		$direction.value = "";

		$tags.innerHTML = "";

		setModel(null);
		setImage(null);

		storeAll();
	});

	$preview.addEventListener("click", () => {
		if (uploading || generating || suggesting) return;

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

		form.append(
			"details",
			prompt("Important details in the image (optional)", "") || "",
		);
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

	$suggest.addEventListener("click", async () => {
		if (uploading || generating | suggesting) return;

		const payload = buildPayload(false);

		if (!payload) {
			return;
		}

		setSuggesting(true);

		let direction = payload.direction;

		$direction.value += `${direction ? "\n\n" : ""}suggesting...`;

		try {
			const response = await fetch("/suggest", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
			});

			if (!response.ok) {
				throw new Error(`Suggestion failed with status ${response.status}`);
			}

			const text = await response.text();

			if (direction) {
				direction += "\n\n";
			}

			direction += text;
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(`${err}`);
			}
		} finally {
			setSuggesting(false);

			direction = clean(dai.clean(direction));

			$direction.value = direction;

			store("direction", direction);
		}
	});

	$tag.addEventListener("keydown", (event) => {
		if (event.key !== "Enter") {
			return;
		}

		event.preventDefault();

		const content = $tag.value.trim();

		if (!content) return;

		appendTag(content);

		$tag.value = "";

		store("tags", buildTags());
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
	setMode(load("mode", null));

	const tags = load("tags", null);

	if (tags && Array.isArray(tags)) {
		for (const tag of tags) {
			appendTag(tag);
		}
	}

	fetch("/models")
		.then((response) => response.json())
		.then(buildModels);
})();
