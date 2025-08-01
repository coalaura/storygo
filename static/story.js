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
		$save = document.getElementById("save"),
		$delete = document.getElementById("delete"),
		$text = document.getElementById("text"),
		$textStatus = document.getElementById("text-status"),
		$mode = document.getElementById("mode"),
		$modeName = document.getElementById("mode-name"),
		$direction = document.getElementById("direction"),
		$directionStatus = document.getElementById("direction-status"),
		$suggest = document.getElementById("suggest"),
		$generate = document.getElementById("generate");

	const $modal = document.getElementById("modal"),
		$mdBackground = $modal.querySelector(".modal-background"),
		$mdTitle = $modal.querySelector(".modal-title"),
		$mdBody = $modal.querySelector(".modal-body"),
		$mdCancel = document.getElementById("modal-cancel"),
		$mdConfirm = document.getElementById("modal-confirm");

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

	function preloadImage(url) {
		return new Promise((resolve) => {
			const img = new Image();

			img.src = url;

			img.onload = () => resolve(img);
			img.onerror = () => resolve(false);
		});
	}

	async function setImage(hash) {
		image = hash;

		if (image) {
			const url = `/i/${image}`;

			$preview.classList.add("image");
			$preview.style.backgroundImage = `url("${url}")`;

			if (!(await preloadImage(url))) {
				setImage(null);
			}
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

	function setTextStatus(status) {
		$textStatus.textContent = status || "";
	}

	function setDirectionStatus(status) {
		$directionStatus.textContent = status || "";
	}

	function setSuggesting(status) {
		suggesting = status;

		if (suggesting) {
			$page.classList.add("suggesting");

			$direction.setAttribute("disabled", "true");
		} else {
			$page.classList.remove("suggesting");

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

	function setModel(set) {
		const models = [...$model.querySelectorAll(".model")];

		if (!models.find((el) => el.dataset.key === set)) {
			set = models[0];
		}

		model = set;

		for (const el of models) {
			if (el.dataset.key === model) {
				el.classList.add("selected");
			} else {
				el.classList.remove("selected");
			}
		}

		store("model", model);
	}

	function buildTags() {
		const tags = [];

		$tags.querySelectorAll(".tag").forEach((tag) => {
			tags.push(tag.textContent.trim());
		});

		return tags;
	}

	function buildNewContext(payload) {
		if (payload.text && payload.image) {
			return "Seamlessly continue the story, maintaining the established tone and plot. Use the provided image as a subtle reference for atmosphere or detail.";
		} else if (payload.text && !payload.image) {
			return "Seamlessly continue the story, maintaining the established tone, characters, and plot.";
		} else if (!payload.text && payload.image) {
			return "Craft a compelling and original story directly inspired by the provided image. Let it be your primary creative anchor.";
		}

		return "Craft a compelling and original story from scratch. You have complete creative freedom.";
	}

	function buildPayload(key, element, inline) {
		const payload = {
			model: model,
			context: clean($context.value),
			text: clean($text.value),
			direction: clean($direction.value),
			tags: buildTags(),
			image: image,
		};

		if (!payload.context) {
			payload.context = buildNewContext(payload);
		}

		if (payload[key]) {
			payload[key] += inline ? " " : "\n\n";
		}

		$context.value = payload.context;
		$text.value = payload.text;
		$direction.value = payload.direction;

		element.scrollTop = element.scrollHeight;

		return payload;
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
	async function stream(url, options, callback) {
		try {
			const response = await fetch(url, options);

			if (!response.ok) {
				throw new Error(`Stream failed with status ${response.status}`);
			}

			const reader = response.body.getReader(),
				decoder = new TextDecoder();

			let buffer = "";

			while (true) {
				const { value, done } = await reader.read();

				if (done) break;

				buffer += decoder.decode(value, {
					stream: true,
				});

				while (true) {
					const idx = buffer.indexOf("\n\n");

					if (idx === -1) {
						break;
					}

					const frame = buffer.slice(0, idx).trim();
					buffer = buffer.slice(idx + 2);

					if (!frame) {
						continue;
					}

					try {
						const chunk = JSON.parse(frame);

						if (!chunk) {
							throw new Error("invalid chunk");
						}

						callback(chunk);
					} catch (err) {
						console.warn("bad frame", frame);
						console.warn(err);
					}
				}
			}
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(`${err}`);
			}
		} finally {
			callback(false);
		}
	}

	let controller;

	async function generate(endpoint, inline) {
		if (uploading || generating || suggesting) return;

		if (endpoint === "overview") {
			$text.value = "";
		}

		let _enable, _status, _element, _key;

		if (endpoint === "suggest") {
			_enable = setSuggesting;
			_status = setDirectionStatus;
			_element = $direction;
			_key = "direction";
		} else {
			_enable = setGenerating;
			_status = setTextStatus;
			_element = $text;
			_key = "text";
		}

		const payload = buildPayload(_key, _element, inline);

		if (!payload) {
			return;
		}

		_enable(true);
		_status("sending");

		controller = new AbortController();

		storeAll();

		stream(
			`/${endpoint}`,
			{
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
				signal: controller.signal,
			},
			(chunk) => {
				if (!chunk) {
					_status(false);

					_enable(false);

					payload[_key] = clean(dai.clean(payload[_key]));

					_element.value = payload[_key];

					store(_key, payload[_key]);

					return;
				}

				if (chunk.state) {
					_status(chunk.state);

					return;
				}

				payload[_key] += chunk.text;

				_element.value = payload[_key];
				_element.scrollTop = _element.scrollHeight;

				store(_key, payload[_key]);

			},
		);
	}

	async function imageExists(file) {
		const form = new FormData();

		form.append("image", file);

		try {
			const response = await fetch("/image/hash", {
				method: "POST",
				body: form,
			});

			if (!response.ok) {
				if (response.status === 404) {
					return false;
				}

				throw new Error(`Check failed with status ${response.status}`);
			}

			setImage(await response.text());

			return true;
		} catch (err) {
			alert(`${err}`);
		}

		return false;
	}

	let mdCallback;

	$mdBackground.addEventListener("click", () => {
		$modal.classList.remove("open");

		mdCallback?.(false);
		mdCallback = null;
	});

	$mdCancel.addEventListener("click", () => {
		$modal.classList.remove("open");

		mdCallback?.(false);
		mdCallback = null;
	});

	$mdConfirm.addEventListener("click", () => {
		$modal.classList.remove("open");

		if (!mdCallback) return;

		const data = {};

		$mdBody.querySelectorAll("input,select").forEach((input) => {
			const name = input.name,
				value = input.value.trim();

			data[name] = value;
		});

		mdCallback(data);
		mdCallback = null;
	});

	function modal(title, html, buttons) {
		mdCallback?.(false);

		return new Promise((resolve) => {
			mdCallback = resolve;

			$mdTitle.textContent = title;
			$mdBody.innerHTML = html;

			$mdBody.querySelectorAll(".dropdown").forEach((dd) => {
				const models = [...dd.querySelectorAll(".model")];

				models[0].classList.add("selected");

				dd.addEventListener("click", (event) => {
					const close = event.target.closest(".model"),
						key = close?.dataset?.key;

					if (!key) return;

					for (const el of models) {
						if (el.dataset.key === key) {
							el.classList.add("selected");
						} else {
							el.classList.remove("selected");
						}
					}
				});
			});

			if (!buttons?.length) {
				buttons = ["Cancel", "Confirm"];
			}

			$mdCancel.textContent = buttons[0];
			$mdConfirm.textContent = buttons[1];

			$modal.classList.add("open");
		});
	}

	async function confirm(title, question) {
		return (await modal(title, `<p>${question}</p>`, ["No", "Yes"])) !== false;
	}

	async function prompt(title, question) {
		const data = await modal(
			title,
			`<p>${question}</p><input type="text" name="prompt" />`,
			["Cancel", "Confirm"],
		);

		if (data === false) {
			return false;
		}

		return data?.prompt || "";
	}

	$model.addEventListener("click", (event) => {
		const close = event.target.closest(".model");

		if (!close) return;

		setModel(close.dataset.key);
	});

	$generate.addEventListener("click", async () => {
		if (generating) {
			controller?.abort?.();

			return;
		}

		generate(mode, false);
	});

	$mode.addEventListener("click", async () => {
		if (generating || suggesting) return;

		if (
			$text.value.trim() &&
			!(await confirm(
				"Switch Modes",
				"Are you sure you want to switch modes? This will clear the story field.",
			))
		) {
			return;
		}

		$text.value = "";

		if (mode === "overview") {
			setMode("generate");
		} else {
			setMode("overview");
		}
	});

	$save.addEventListener("click", async () => {
		if (generating || uploading) return;

		const text = clean($text.value);

		if (!text) {
			alert("Story is empty");

			return;
		}

		const doc = new jspdf.jsPDF({
			orientation: "portrait",
			unit: "mm",
			format: "a4",
		});

		const docWidth = doc.internal.pageSize.width,
			docHeight = doc.internal.pageSize.height;

		const margin = 15,
			lineHeight = docHeight / 62,
			gutter = 4,
			maxWidth = docWidth - margin * 2;

		doc.setFont("Times", "Roman");
		doc.setFontSize(12);

		let imageWidth = 0,
			imageHeight = 0,
			imageBottomY = 0;

		if (image) {
			const img = await preloadImage(`/i/${image}`);

			if (img) {
				const aspect = img.height / img.width;

				imageWidth = 60;
				imageHeight = imageWidth * aspect;

				imageBottomY = margin + imageHeight;

				doc.addImage(
					img,
					"PNG",
					docWidth - margin - imageWidth,
					margin,
					imageWidth,
					imageHeight,
				);
			}
		}

		let y = margin,
			pages = 1,
			passed;

		function renderFooter() {
			const footerY = docHeight - margin + lineHeight / 2;

			doc.text(pages.toString(), docWidth - margin, footerY, {
				baseline: "top",
				align: "right",
			});
		}

		renderFooter();

		const paragraphs = text.split(/(\r?\n)+/g).map((p) => p.trim());

		for (const paragraph of paragraphs) {
			let remainingText = paragraph;

			while (remainingText.length > 0) {
				let section = [],
					nextY = y,
					textWidth = maxWidth;

				while (remainingText.length > 0) {
					if (nextY + lineHeight > docHeight - margin) {
						doc.addPage();
						pages++;

						renderFooter();

						y = margin;
						nextY = margin;
					}

					let newTextWidth = maxWidth;

					if (imageHeight > 0 && nextY < imageBottomY) {
						newTextWidth = maxWidth - imageWidth - gutter;
					} else if (!passed) {
						passed = true;

						break;
					}

					textWidth = newTextWidth;

					const lines = doc.splitTextToSize(remainingText, textWidth),
						line = lines[0];

					section.push(line);

					nextY += lineHeight;

					remainingText = remainingText.substring(line.length).trim();
				}

				doc.text(section.join("\n"), margin, y, {
					baseline: "top",
					align: "justify",
					maxWidth: textWidth,
				});

				y += section.length * lineHeight;
			}

			y += lineHeight / 2;
		}

		doc.save("story.pdf");
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

	$delete.addEventListener("click", async () => {
		if (uploading || generating || suggesting) return;

		if (
			!(await confirm(
				"Clear Data",
				"Are you sure you want to clear the story, context, directions and image?",
			))
		) {
			return;
		}

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

		if (await imageExists(file)) {
			setUploading(false);

			return;
		}

		const details = await prompt(
			"Image Details",
			"Important details in the image, that the image-to-text model should pay close attention to or might miss. (optional)",
		);

		if (details === false) {
			setUploading(false);

			return;
		}

		const form = new FormData();

		form.append("image", file);
		form.append("details", details);

		let hash;

		try {
			const response = await fetch("/image/upload", {
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
		if (suggesting) {
			controller?.abort?.();

			return;
		}

		generate("suggest", false);
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

			generate(mode, true);
		}
	});

	$direction.addEventListener("change", () => {
		store("direction", clean($direction.value));
	});

	$direction.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "Enter") {
			event.preventDefault();

			$generate.click();
		} else if (event.key === "Tab") {
			event.preventDefault();

			generate("suggest", true);
		}
	});

	document.addEventListener("keydown", (event) => {
		if (event.ctrlKey && event.key === "s") {
			event.preventDefault();

			$save.click();
		}
	});

	$context.value = load("context", "");
	$text.value = load("text", "");
	$direction.value = load("direction", "");

	$text.scrollTop = $text.scrollHeight;

	setModel(load("model", null));
	setImage(load("image", null));
	setMode(load("mode", null));

	const tags = load("tags", null);

	if (tags && Array.isArray(tags)) {
		for (const tag of tags) {
			appendTag(tag);
		}
	}
})();
