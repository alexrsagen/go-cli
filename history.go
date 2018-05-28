package cli

type line struct {
	original, edited string
	isEdited         bool
}

type history struct {
	entries []*line
	index   int
}

func (h *history) isNew() bool {
	return len(h.entries) == 0 || h.index == len(h.entries)
}

func (h *history) isFirst() bool {
	return len(h.entries) == 0 || h.index == 0
}

func (h *history) isLast() bool {
	return len(h.entries) == 0 || h.index == len(h.entries)-1
}

func (h *history) first() bool {
	if h.isFirst() {
		return false
	}
	h.index = 0
	return true
}

func (h *history) last() bool {
	if h.isLast() {
		return false
	}
	h.index = len(h.entries) - 1
	return true
}

func (h *history) prev() bool {
	if h.isFirst() {
		return false
	}
	h.index--
	return true
}

func (h *history) next() bool {
	if h.isLast() {
		return false
	}
	h.index++
	return true
}

func (h *history) new() {
	h.last()
	if h.get() == "" {
		return
	}
	h.index++
	h.entries = append(h.entries, &line{})
}

func (h *history) get() string {
	if h.isNew() {
		return ""
	}
	if h.entries[h.index].isEdited {
		return h.entries[h.index].edited
	}
	return h.entries[h.index].original
}

func (h *history) set(s string) {
	// If last history entry, store changes in original, otherwise store in edited
	if h.isNew() {
		h.entries = append(h.entries, &line{original: s})
		return
	} else if h.isLast() {
		h.entries[h.index].original = s
	} else if s == h.entries[h.index].original {
		h.entries[h.index].isEdited = false
	} else {
		h.entries[h.index].edited = s
		h.entries[h.index].isEdited = true
	}
}

func (h *history) revert() {
	if h.isNew() || !h.entries[h.index].isEdited {
		return
	}
	h.entries[h.index].edited = ""
	h.entries[h.index].isEdited = false
}

func (h *history) revertAndAdd() {
	if h.isLast() || !h.entries[h.index].isEdited {
		return
	}
	h.entries[len(h.entries)-1] = &line{original: h.entries[h.index].edited}
	h.revert()
}
