#!/bin/bash
# differ.sh: Generate LaTeX differences between two Sphinx build directories.

set -eu

old="$1/latex/"
new="$2/latex/"

shopt -s nullglob
for oldfile in ${old}*.tex; do
  basename="$(basename "${oldfile}" .tex)"
  newfile="${new}${basename}.tex"
  difffile="${new}${basename}-diff.tex"
  if [[ "${basename}" != "*-diff" && -f "${newfile}" ]]; then
    latexdiff --replace-context2cmd="none" --config VERBATIMENV=sphinxVerbatim "${oldfile}" "${newfile}" > "${difffile}"

    # Only generate the -diff.pdf if there are any actual differences.
    if grep --invert-match "%DIF PREAMBLE" "${difffile}" | grep --quiet '\\DIF'; then
      # latexdiff can mess up latex sometimes, but output is still visually fine
      make --directory "${new}" "${basename}-diff.pdf" || true
    else
      # Remove any existing -diff.pdf to signal that there were no changes.
      rm --force "${new}${basename}-diff.pdf"
    fi
  fi
done

exit 0
