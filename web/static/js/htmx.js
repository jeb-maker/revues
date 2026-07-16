/* Minimal HTMX client — subset of htmx.org within the 15 Ko eco budget. */
(function () {
  "use strict";

  var baseHeaders = { "HX-Request": "true" };

  function csrfToken() {
    var meta = document.querySelector('meta[name="csrf-token"]');
    return meta ? meta.getAttribute("content") : "";
  }

  function associatedWithForm(el, form) {
    if (!el || !form) {
      return false;
    }
    if (form.contains(el)) {
      return true;
    }
    return form.id && el.getAttribute("form") === form.id;
  }

  function parseTriggers(raw) {
    return raw.split(",").map(function (part) {
      part = part.trim();
      var event = part.split(/\s+/)[0];
      var fromMatch = part.match(/from:([^\s]+)/);
      var keyMatch = part.match(/\[key==['"](.+?)['"]\]/);
      var needsCtrl = part.indexOf("ctrlKey") !== -1;
      return {
        event: event,
        from: fromMatch ? fromMatch[1] : null,
        key: keyMatch ? keyMatch[1] : null,
        ctrlKey: needsCtrl,
      };
    });
  }

  function appendField(fd, el) {
    if (!el.name || el.disabled) {
      return;
    }
    if ((el.type === "checkbox" || el.type === "radio") && !el.checked) {
      return;
    }
    if (el.tagName === "SELECT") {
      if (el.multiple) {
        Array.prototype.forEach.call(el.selectedOptions, function (opt) {
          fd.append(el.name, opt.value);
        });
      } else {
        fd.append(el.name, el.value);
      }
      return;
    }
    fd.append(el.name, el.value);
  }

  function formDataFor(form) {
    var fd = new FormData();
    form.querySelectorAll("input,select,textarea").forEach(function (el) {
      appendField(fd, el);
    });
    if (form.id) {
      document.querySelectorAll('[form="' + form.id + '"]').forEach(function (el) {
        appendField(fd, el);
      });
    }
    return fd;
  }

  function swapTarget(target, html, mode) {
    if (!target) {
      return null;
    }
    if (mode === "outerHTML") {
      target.outerHTML = html;
      return document.getElementById(target.id);
    }
    target.innerHTML = html;
    return target;
  }

  function applyOOB(container) {
    container.querySelectorAll("[hx-swap-oob]").forEach(function (el) {
      if (!el.id) {
        return;
      }
      var existing = document.getElementById(el.id);
      if (!existing) {
        return;
      }
      var mode = el.getAttribute("hx-swap-oob") || "true";
      swapTarget(existing, el.outerHTML, mode === "true" ? "outerHTML" : mode);
    });
  }

  function request(form, triggerEl) {
    var url = form.getAttribute("hx-post") || form.getAttribute("hx-get");
    if (!url) {
      return;
    }
    var confirmMsg = (triggerEl && triggerEl.getAttribute("hx-confirm")) || form.getAttribute("hx-confirm");
    if (confirmMsg && !window.confirm(confirmMsg)) {
      if (triggerEl && triggerEl.tagName === "SELECT") {
        triggerEl.dispatchEvent(new Event("htmx:confirm:cancelled"));
      }
      return;
    }
    var method = form.getAttribute("hx-post") ? "POST" : "GET";
    var targetSel = form.getAttribute("hx-target");
    var swapMode = form.getAttribute("hx-swap") || "innerHTML";
    var target = targetSel ? document.querySelector(targetSel) : form;
    var indicatorSel = (triggerEl && triggerEl.getAttribute("hx-indicator")) || form.getAttribute("hx-indicator");
    var indicator = indicatorSel ? document.querySelector(indicatorSel) : null;

    var headers = Object.assign({}, baseHeaders);
    var extra = form.getAttribute("hx-headers");
    if (extra) {
      try {
        Object.assign(headers, JSON.parse(extra));
      } catch (e) {
        /* ignore malformed hx-headers */
      }
    }
    var token = csrfToken();
    if (token) {
      headers["X-CSRF-Token"] = token;
    }

    if (indicator) {
      indicator.classList.add("htmx-request");
    }
    var body = method === "POST" ? formDataFor(form) : null;
    fetch(url, { method: method, headers: headers, body: body, credentials: "same-origin" })
      .then(function (resp) {
        return resp.text().then(function (text) {
          return { ok: resp.ok, text: text };
        });
      })
      .then(function (result) {
        var wrapper = document.createElement("div");
        wrapper.innerHTML = result.text;
        applyOOB(wrapper);
        if (target) {
          var main = wrapper.firstElementChild;
          if (main && !main.hasAttribute("hx-swap-oob")) {
            swapTarget(target, main.outerHTML, swapMode);
          }
        }
        process(document);
      })
      .finally(function () {
        if (indicator) {
          indicator.classList.remove("htmx-request");
        }
      });
  }

  function matchesTrigger(form, event) {
    var trigger = form.getAttribute("hx-trigger") || "submit";
    return parseTriggers(trigger).some(function (spec) {
      if (spec.event !== event.type) {
        return false;
      }
      if (spec.from && !event.target.matches(spec.from)) {
        return false;
      }
      if (spec.key && event.key !== spec.key) {
        return false;
      }
      if (spec.ctrlKey && !event.ctrlKey) {
        return false;
      }
      return associatedWithForm(event.target, form);
    });
  }

  function bindForm(form) {
    if (form.dataset.hxBound === "1") {
      return;
    }
    form.dataset.hxBound = "1";

    form.addEventListener("submit", function (event) {
      if (!form.getAttribute("hx-post")) {
        return;
      }
      event.preventDefault();
      request(form, event.submitter || null);
    });
  }

  function handleDocumentEvent(event) {
    document.querySelectorAll("form[hx-post],form[hx-get]").forEach(function (form) {
      if (!matchesTrigger(form, event)) {
        return;
      }
      if (event.type === "submit") {
        event.preventDefault();
      }
      request(form, event.target);
    });
  }

  function process(root) {
    root.querySelectorAll("form[hx-post],form[hx-get]").forEach(bindForm);
  }

  window.htmx = { process: process };
  document.addEventListener("DOMContentLoaded", function () {
    process(document);
    ["change", "blur", "keydown"].forEach(function (name) {
      document.addEventListener(name, handleDocumentEvent);
    });
  });
})();
