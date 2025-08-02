function make(tag, classes) {
	const el = document.createElement(tag);

	el.classList.add(...classes);

	return el;
}

async function get(url, type = false) {
	try {
		const response = await fetch(url);

		if (!response.ok) {
			throw new Error(response.statusText);
		}

		switch (type) {
			case "text":
				return await response.text();
			case "json":
				return await response.json();
			case "blob":
				return await response.blob();
		}

		return true;
	} catch (err) {
		console.warn(err);
	}

	return false;
}

(() => {
	const $page = document.getElementById("page"),
		$model = document.getElementById("model"),
		$context = document.getElementById("context"),
		$preview = document.getElementById("preview"),
		$image = document.getElementById("image"),
		$createTags = document.getElementById("create-tags"),
		$tag = document.getElementById("tag"),
		$tags = document.getElementById("tags"),
		$createImage = document.getElementById("create-image"),
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
		tagging,
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

	function setValue(el, value) {
		el.value = value;

		el.dispatchEvent(new Event("input"));
		el.dispatchEvent(new Event("change"));
	}

	async function setImage(hash) {
		image = hash;

		if (image) {
			const url = `/i/${image}`;

			$preview.classList.add("image");
			$preview.style.backgroundImage = `url("${url}")`;

			if (!(await get(url))) {
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

	function setTagging(status) {
		tagging = status;

		if (tagging) {
			$page.classList.add("tagging");

			$tag.setAttribute("disabled", "true");
		} else {
			$page.classList.remove("tagging");

			$tag.removeAttribute("disabled");
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
			if (tagging) return;

			tag.remove();

			store("tags", buildTags());
		});

		$tags.appendChild(tag);
	}

	function setModel(set) {
		const models = [...$model.querySelectorAll(".model")];

		if (!set || !models.find((el) => el.dataset.key === set)) {
			set = models[0].dataset.key;
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

		if (key && payload[key]) {
			payload[key] += inline ? " " : "\n\n";
		}

		setValue($context, payload.context);
		setValue($text, payload.text);
		setValue($direction, payload.direction);

		if (element) {
			element.scrollTop = element.scrollHeight;
		}

		return payload;
	}

	function download(name, type, data) {
		let blob;

		if (data instanceof Blob) {
			blob = data;
		} else {
			blob = new Blob([data], {
				type: type,
			});
		}

		const a = document.createElement("a"),
			url = URL.createObjectURL(blob);

		a.setAttribute("download", name);
		a.style.display = "none";
		a.href = url;

		document.body.appendChild(a);

		a.click();

		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	async function downloadImage(name, url, type) {
		try {
			const response = await fetch(url);

			if (!response.ok) {
				throw new Error(response.statusText);
			}

			download(name, type, await response.blob());
		} catch (err) {
			console.error(err);
		}
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
				alert(err.message);
			}
		} finally {
			callback(false);
		}
	}

	let controller;

	async function generate(endpoint, inline) {
		if (uploading || generating || suggesting) return;

		if (endpoint === "overview") {
			setValue($text, "");
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

					setValue(_element, payload[_key]);

					store(_key, payload[_key]);

					return;
				}

				if (chunk.state) {
					_status(chunk.state);

					return;
				}

				payload[_key] += chunk.text || chunk.error;

				setValue(_element, payload[_key]);
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

	async function uploadImage(file, allowDetails = false) {
		setUploading(true);

		if (await imageExists(file)) {
			setUploading(false);

			return;
		}

		const form = new FormData();

		form.append("image", file);

		if (allowDetails) {
			const details = await prompt(
				"Image Details",
				"Important details in the image, that the image-to-text model should pay close attention to or might miss. (optional)",
			);

			if (details === false) {
				setUploading(false);

				return;
			}

			form.append("details", details);
		}

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
			alert(err.message);

			return;
		} finally {
			setUploading(false);
		}

		setImage(hash);
	}

	function alert(message) {
		const wrapper = make("div", ["alert"]),
			back = make("div", ["alert-background"]),
			cont = make("div", ["alert-content"]),
			body = make("div", ["alert-body"]),
			btns = make("div", ["alert-buttons"]),
			conf = make("button", ["alert-confirm"]);

		body.textContent = message;
		conf.textContent = "Okay";

		wrapper.appendChild(back);
		wrapper.appendChild(cont);

		cont.appendChild(body);
		cont.appendChild(btns);

		btns.appendChild(conf);

		back.addEventListener("click", () => {
			wrapper.remove();
		});

		conf.addEventListener("click", () => {
			wrapper.remove();
		});

		document.body.appendChild(wrapper);
	}

	let mdCallback;

	function closeModal() {
		$modal.classList.remove("open");

		mdCallback?.(false);
		mdCallback = null;
	}

	$mdBackground.addEventListener("click", () => {
		closeModal();
	});

	$mdCancel.addEventListener("click", () => {
		closeModal();
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

		$mdBody.querySelectorAll(".dropdown").forEach((input) => {
			const name = input.dataset.name,
				selected = input.querySelector(".model.selected");

			if (!name) {
				return;
			}

			data[name] = selected?.dataset?.key;
		});

		mdCallback(data);
		mdCallback = null;
	});

	function dropdown(name, options) {
		const optionsHtml = options
			.map((option) => {
				let tags = "";

				if (option.tags) {
					const list = option.tags.map(
						(tag) =>
							`<div class="tag" title="${tag}" style="background-image:url('/icons/tags/${tag}.svg')"></div>`,
					);

					tags = `<div class="tags">${list.join("")}</div>`;
				}

				return `<div class="model" data-key="${option.key}"><div class="name ${option.vision ? "vision" : ""}">${option.name}</div>${tags}</div>`;
			})
			.join("");

		return `<div data-name="${name}" class="dropdown"><div class="options">${optionsHtml}</div></div>`;
	}

	function modal(title, html, buttons) {
		closeModal();

		$mdTitle.textContent = title;
		$mdBody.innerHTML = html;

		return new Promise((resolve) => {
			mdCallback = resolve;

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

			if (buttons?.length && buttons?.length >= 1) {
				$mdCancel.classList.remove("hidden");

				$mdCancel.textContent = buttons[0];
			} else {
				$mdCancel.classList.add("hidden");
			}

			if (buttons?.length === 2) {
				$mdConfirm.classList.remove("hidden");

				$mdConfirm.textContent = buttons[1];
			} else {
				$mdConfirm.classList.add("hidden");
			}

			$modal.classList.add("open");
		});
	}

	async function confirm(title, question) {
		return (await modal(title, `<p>${question}</p>`, ["No", "Yes"])) !== false;
	}

	async function prompt(title, question, value = "") {
		const data = await modal(
			title,
			`<p>${question}</p><input type="text" name="prompt" value="${value}" />`,
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

		setValue($text, "");

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

		download("story.txt", "text/plain", text);

		if (image) {
			downloadImage("story.webp", `/i/${image}`, "image/webp");
		}
	});

	$export.addEventListener("click", async () => {
		if (generating || suggesting) return;

		const zip = new JSZip();

		const payload = buildPayload(false, false, false),
			metadata = {
				mode: mode,
				model: model,
				tags: buildTags(),
			};

		// store metadata
		zip.file(
			"metadata.json",
			new Blob([JSON.stringify(metadata)], {
				type: "application/json",
			}),
		);

		// store context
		if (payload.context) {
			zip.file(
				"context.txt",
				new Blob([payload.context], {
					type: "text/plain",
				}),
			);
		}

		// store text
		if (payload.text) {
			zip.file(
				"text.txt",
				new Blob([payload.text], {
					type: "text/plain",
				}),
			);
		}

		// store direction
		if (payload.direction) {
			zip.file(
				"direction.txt",
				new Blob([payload.direction], {
					type: "text/plain",
				}),
			);
		}

		// store image
		if (image) {
			const img = await get(`/i/${image}`, "blob");

			if (img) {
				zip.file("image.webp", img);
			} else {
				alert("Failed to load image.");
			}
		}

		// build & download zip
		const blob = await zip.generateAsync({
			type: "blob",
		});

		download("story.zip", "application/zip", blob);
	});

	$import.addEventListener("click", () => {
		if (generating || uploading || suggesting) return;

		$importFile.click();
	});

	$importFile.addEventListener("change", (event) => {
		const file = event.target.files[0];

		if (!file) return;

		const reader = new FileReader();

		reader.onload = async (event) => {
			try {
				const zip = await JSZip.loadAsync(event.target.result);

				// read metadata
				let file = zip.file("metadata.json");

				const metadata = file ? JSON.parse(await file.async("string")) : null;

				if (metadata?.model && typeof metadata.model === "string") {
					setModel(metadata.model);
				} else {
					setModel(null);
				}

				$tags.innerHTML = "";

				if (metadata?.tags && Array.isArray(metadata.tags)) {
					for (const tag of metadata.tags) {
						appendTag(tag);
					}
				}

				if (metadata?.mode && typeof metadata.mode === "string") {
					setMode(metadata.mode);
				} else {
					setMode("generate");
				}

				// read context
				file = zip.file("context.txt");

				if (file) {
					setValue($context, clean(await file.async("string")));
				} else {
					setValue($context, "");
				}

				// read text
				file = zip.file("text.txt");

				if (file) {
					setValue($text, clean(await file.async("string")));

					$text.scrollTop = $text.scrollHeight;
				} else {
					setValue($text, "");
				}

				// read direction
				file = zip.file("direction.txt");

				if (file) {
					setValue($direction, clean(await file.async("string")));
				} else {
					setValue($direction, "");
				}

				// read image
				file = zip.file("image.webp");

				if (file) {
					uploadImage(await file.async("blob"));
				} else {
					setImage(null);
				}

				storeAll();
			} catch (err) {
				console.error(err);

				alert("Invalid safe file.");
			}
		};

		reader.readAsArrayBuffer(file);
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

		setValue($context, "");
		setValue($text, "");
		setValue($direction, "");

		$tags.innerHTML = "";

		setModel(null);
		setImage(null);

		storeAll();
	});

	$createImage.addEventListener("click", async () => {
		if (uploading || generating || suggesting) return;

		const modelDropdown = dropdown("model", ImageModels),
			styleDropdown = dropdown(
				"style",
				ImageStyles.map((style, index) => ({
					key: index,
					name: style,
				})),
			);

		const selection = await modal(
			"Image Generation",
			`<div class="form-group">
				<label>Image Model</label>
				${modelDropdown}
			</div>
			<div class="form-group">
				<label>Image Model</label>
				${styleDropdown}
			</div>`,
			["Cancel", "Generate"],
		);

		if (!selection) {
			return;
		}

		const payload = buildPayload(false, false, false);

		payload.model = selection.model;

		const abort = new AbortController();

		let resultName, resultUrl;

		modal("Image Generation", `<p>Waiting...</p>`, ["Abort"]).then(
			async (shouldSave) => {
				abort.abort();

				if (shouldSave && resultUrl) {
					downloadImage(resultName, resultUrl, "image/webp");
				}
			},
		);

		const $mInfo = $mdBody.querySelector("p");

		stream(
			`/image/create/${selection.style || 0}`,
			{
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
				signal: abort.signal,
			},
			(chunk) => {
				if (!chunk) {
					if (!resultUrl) {
						closeModal();
					}

					return;
				}

				if (chunk.error) {
					alert(chunk.error);

					return;
				} else if (chunk.state) {
					switch (chunk.state) {
						case "prompt":
							$mInfo.textContent =
								"Creating image prompt, this may take a moment...";
							break;
						case "image":
							$mInfo.textContent =
								"Generating image, this may take a moment...";
							break;
						case "save":
							$mInfo.textContent = "Saving generated image...";
							break;
					}

					return;
				}

				resultName = `${chunk.text}.webp`;
				resultUrl = `/g/${chunk.text}`;

				$mdBody.innerHTML = `<img src="${resultUrl}" />`;

				$mdCancel.textContent = "Close";
				$mdConfirm.textContent = "Download";

				$mdConfirm.classList.remove("hidden");
			},
		);
	});

	$preview.addEventListener("click", () => {
		if (uploading || generating || suggesting) return;

		if (image) {
			setImage(null);

			return;
		}

		$image.click();
	});

	$image.addEventListener("change", (event) => {
		const file = event.target.files[0];

		if (!file) {
			return;
		}

		uploadImage(file);
	});

	$suggest.addEventListener("click", async () => {
		if (suggesting) {
			controller?.abort?.();

			return;
		}

		generate("suggest", false);
	});

	let tagController, tagNotes;

	$createTags.addEventListener("click", async () => {
		if (tagging) {
			tagController?.abort();

			return;
		}

		tagController = new AbortController();

		setTagging(true);

		tagNotes =
			(await prompt(
				"Tag Notes",
				"Any additional notes to keep in mind when creating the tags. (optional)",
				tagNotes || "",
			)) || "";

		const payload = buildPayload(false, false, false);

		payload.notes = tagNotes;

		try {
			const response = await fetch("/tags", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
				signal: tagController.signal,
			});

			if (!response.ok) {
				throw new Error(`Tagging failed with status ${response.status}`);
			}

			const list = await response.json();

			if (!list || !Array.isArray(list)) {
				throw new Error("Invalid tag list returned");
			}

			$tags.innerHTML = "";

			for (const tag of list) {
				appendTag(tag);
			}

			store("tags", buildTags());
		} catch (err) {
			if (err.name !== "AbortError") {
				alert(err.message);
			}
		} finally {
			setTagging(false);
		}
	});

	$tag.addEventListener("keydown", (event) => {
		if (tagging || event.key !== "Enter") {
			return;
		}

		event.preventDefault();

		const content = $tag.value.trim();

		if (!content) return;

		appendTag(content);

		setValue($tag, "");

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

	$text.addEventListener("input", () => {
		if (clean($text.value)) {
			$page.classList.add("not-empty");
		} else {
			$page.classList.remove("not-empty");
		}
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

	setValue($context, load("context", ""));
	setValue($text, load("text", ""));
	setValue($direction, load("direction", ""));

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
