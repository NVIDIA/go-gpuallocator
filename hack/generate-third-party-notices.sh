#!/usr/bin/env bash
# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

OUTPUT="${OUTPUT:-THIRD_PARTY_NOTICES.md}"
MODULES_TXT="${MODULES_TXT:-vendor/modules.txt}"

PACKAGES=("./...")

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
)

# CGO must stay on: with it off, go-nvml/pkg/dl leaves the import graph and
# go-licenses reports go-nvml as ".../pkg/nvml" rather than ".../pkg".
export CGO_ENABLED=1

die() {
    printf 'ERROR: %s\n' "$1" >&2
    shift
    if (( $# > 0 )); then
        printf '%s\n' "$@" >&2
    fi
    exit 1
}

log() {
    printf '%s\n' "$*" >&2
}

# A license that is itself Markdown would close a fixed ``` fence early.
fence_for() {
    local file="$1" longest_backtick_run fence_width
    # -a: a license holding a NUL byte would print "Binary file ... matches".
    longest_backtick_run=$(LC_ALL=C grep -oaE '`+' "${file}" 2>/dev/null \
        | awk '{ if (length($0) > longest) longest = length($0) } END { print longest+0 }')
    fence_width=$(( longest_backtick_run + 1 ))
    (( fence_width < 3 )) && fence_width=3
    printf '%*s' "${fence_width}" '' | tr ' ' '`'
}

check_prerequisites() {
    command -v go >/dev/null 2>&1 || die "go is not installed."

    if ./bin/go-licenses --help >/dev/null 2>&1; then
        GO_LICENSES="${PWD}/bin/go-licenses"
    elif command -v go-licenses >/dev/null 2>&1; then
        GO_LICENSES="$(command -v go-licenses)"
    else
        die "no usable go-licenses binary found." \
            "Run 'rm -f bin/go-licenses': a binary built for another platform cannot run here," \
            "and make will not replace one that already exists. Make reinstalls it once it is gone."
    fi

    [[ -f "${MODULES_TXT}" ]] \
        || die "${MODULES_TXT} not found — run this from the repo root, or 'make vendor' to create it."
    [[ -r "${MODULES_TXT}" ]] \
        || die "${MODULES_TXT} is not readable."

    LOCAL_MODULE=$(go list -m 2>/dev/null || true)
    [[ -n "${LOCAL_MODULE}" ]] || die "could not determine local module path via 'go list -m'."

    export GOFLAGS="-mod=vendor"
}

prepare_workspace() {
    # Explicit templates: macOS mktemp ignores TMPDIR without one.
    local work_dir_template="${TMPDIR:-/tmp}/go-gpuallocator-notices"
    WORK_DIR="$(mktemp -d "${work_dir_template}.XXXXXX")"
    SAVE_ROOT="${WORK_DIR}/save"
    LICENSES_DIR="${WORK_DIR}/licenses"
    COMBINED_CSV="${WORK_DIR}/licenses.csv"
    INDEX_FILE="${WORK_DIR}/index"
    mkdir -p "${SAVE_ROOT}" "${LICENSES_DIR}"
    : > "${COMBINED_CSV}"

    local output_dir
    output_dir="$(dirname "${OUTPUT}")"
    mkdir -p "${output_dir}"
    OUTPUT_TMP="$(mktemp "${output_dir}/.$(basename "${OUTPUT}").XXXXXX")"
    trap 'rm -rf "${WORK_DIR}"; rm -f "${OUTPUT_TMP}"' EXIT
}

collect_licenses() {
    local platform goos goarch platform_save_dir

    for platform in "${PLATFORMS[@]}"; do
        goos="${platform%/*}"
        goarch="${platform#*/}"
        log "Collecting licenses for ${goos}/${goarch}..."

        platform_save_dir="${SAVE_ROOT}/${goos}_${goarch}"

        # Only the local module: --ignore matches raw string prefixes, so a
        # stdlib list adds the token "go" and drops golang.org/x/*, gopkg.in/*.
        GOOS="${goos}" GOARCH="${goarch}" "${GO_LICENSES}" save "${PACKAGES[@]}" \
            --save_path="${platform_save_dir}" \
            --force \
            --ignore="${LOCAL_MODULE}"

        GOOS="${goos}" GOARCH="${goarch}" "${GO_LICENSES}" csv "${PACKAGES[@]}" \
            --ignore="${LOCAL_MODULE}" \
            >> "${COMBINED_CSV}"

        cp -R "${platform_save_dir}/." "${LICENSES_DIR}/"
        chmod -R u+w "${LICENSES_DIR}"
    done
}

# Whole-line sort: key-only dedup drops rows, and differs between BSD and GNU.
collapse_index() {
    LC_ALL=C sort -u "$1" | awk -F, '
        {
            package_path = $1
            if (!(package_path in source_url)) {
                source_url[package_path] = $2
                package_order[++package_count] = package_path
            }
            if (!((package_path SUBSEP $3) in seen_license)) {
                seen_license[package_path SUBSEP $3] = 1
                # Count, do not test "package_path in joined_licenses": mawk
                # instantiates the target before evaluating the RHS, BWK awk does not.
                joined_licenses[package_path] = \
                    (license_count[package_path]++ ? joined_licenses[package_path] " / " : "") $3
            }
        }
        END {
            for (i = 1; i <= package_count; i++) {
                package_path = package_order[i]
                print package_path "," source_url[package_path] "," joined_licenses[package_path]
            }
        }
    '
}

# module@version, not a URL: in vendor mode go-licenses points into this repo.
annotate_modules() {
    awk -v modules_txt="${MODULES_TXT}" '
        BEGIN {
            FS = OFS = ","
            while ((getline line < modules_txt) > 0) {
                if (line !~ /^# /) continue
                split(line, fields, " ")
                if (fields[4] == "=>" || fields[3] == "=>") {
                    replacement_field = (fields[4] == "=>") ? 5 : 4
                    if (fields[replacement_field + 1] == "") {
                        print "ERROR: " modules_txt " replaces " fields[2] " with a local path;" > "/dev/stderr"
                        print "teach hack/generate-third-party-notices.sh how to attribute it." > "/dev/stderr"
                        exit 1
                    }
                    module_paths[++module_count] = fields[2]
                    module_display[fields[2]] = fields[replacement_field] "@" fields[replacement_field + 1]
                } else {
                    module_paths[++module_count] = fields[2]
                    module_display[fields[2]] = fields[2] "@" fields[3]
                }
            }
            close(modules_txt)
            if (module_count == 0) {
                print "ERROR: no module lines read from " modules_txt > "/dev/stderr"
                exit 1
            }
        }
        {
            longest_match = ""
            for (i = 1; i <= module_count; i++) {
                module_path = module_paths[i]
                if (($1 == module_path || index($1, module_path "/") == 1) \
                    && length(module_path) > length(longest_match)) {
                    longest_match = module_path
                }
            }
            print $0, (longest_match == "" ? "unknown" : module_display[longest_match])
        }
    '
}

build_index() {
    log "Generating dependency index..."
    collapse_index "${COMBINED_CSV}" | annotate_modules > "${INDEX_FILE}"

    [[ -s "${INDEX_FILE}" ]] \
        || die "go-licenses produced no entries for ${PACKAGES[*]} — refusing to write empty notices file."

    # go-licenses reports an unclassifiable license as "Unknown" and exits 0.
    if cut -d, -f3 "${INDEX_FILE}" | LC_ALL=C grep -qE '^$|(^| / )Unknown( / |$)'; then
        die "go-licenses could not classify the license of some packages." \
            "Inspect them by hand rather than committing a file that says 'Unknown'."
    fi

    if cut -d, -f4 "${INDEX_FILE}" | LC_ALL=C grep -qx 'unknown'; then
        die "could not resolve module@version for some packages from ${MODULES_TXT}." \
            "Run 'make vendor' and re-run, rather than committing a file with unattributed entries."
    fi
}

# Filter by name: for restricted licenses 'go-licenses save' copies whole source.
license_files_for() {
    local package_dir="$1" candidate_file
    [[ -d "${package_dir}" ]] || return 0
    while IFS= read -r -d '' candidate_file; do
        if printf '%s' "$(basename "${candidate_file}")" \
            | LC_ALL=C grep -qiE '^(licen[cs]e|notice|copying|copyright|authors|patents)([-._].*)?$'; then
            printf '%s\n' "${candidate_file}"
        fi
    done < <(find "${package_dir}" -maxdepth 1 -type f -print0 2>/dev/null | LC_ALL=C sort -z)
}

emit_index_table() {
    local package_path _source_url license module
    printf '| Package | License | Module |\n'
    printf '|---------|---------|--------|\n'

    while IFS=, read -r package_path _source_url license module; do
        [[ -z "${package_path}" ]] && continue
        # shellcheck disable=SC2016  # backticks are literal markdown here.
        printf '| `%s` | %s | `%s` |\n' "${package_path}" "${license:-Unknown}" "${module:-unknown}"
    done < "${INDEX_FILE}"
}

emit_sections() {
    local package_path _source_url license module license_files license_file fence

    while IFS=, read -r package_path _source_url license module; do
        [[ -z "${package_path}" ]] && continue

        printf '### %s\n\n' "${package_path}"
        printf '* License: %s\n' "${license:-Unknown}"
        printf '* Module: %s\n\n' "${module:-unknown}"

        license_files=()
        while IFS= read -r license_file; do
            [[ -n "${license_file}" ]] && license_files+=("${license_file}")
        done < <(license_files_for "${LICENSES_DIR}/${package_path}")

        if (( ${#license_files[@]} == 0 )); then
            printf 'License text unavailable. See upstream source for the full license.\n'
        else
            for license_file in "${license_files[@]}"; do
                fence="$(fence_for "${license_file}")"
                printf '#### %s\n\n' "$(basename "${license_file}")"
                printf '%stext\n' "${fence}"
                cat "${license_file}"
                echo
                printf '%s\n' "${fence}"
                echo
            done
        fi
        echo
    done < "${INDEX_FILE}"
}

compose_document() {
    log "Composing ${OUTPUT}..."
    {
        cat <<'EOF'
# Third-Party Notices

NVIDIA go-gpuallocator

This file lists the third-party Go modules that `go-gpuallocator` links into the
packages a consumer imports, along with the verbatim text of each dependency's
license. `go-gpuallocator` is a library: the module itself is the unit of
distribution, so this covers every package in it.

Go standard library packages are excluded; they are covered by the license of
the Go distribution itself. Dependencies reached only from `_test.go` files are
excluded; they are vendored for testing but a consumer does not link them.

## Dependency Index

EOF
        emit_index_table

        cat <<'EOF'

## License Texts

EOF
        emit_sections
    } > "${OUTPUT_TMP}"

    # mktemp creates 0600, and mv within OUTPUT's directory is an atomic rename.
    chmod 644 "${OUTPUT_TMP}"
    mv -f "${OUTPUT_TMP}" "${OUTPUT}"
}

main() {
    check_prerequisites
    prepare_workspace

    collect_licenses
    build_index
    compose_document

    local package_count
    package_count=$(wc -l < "${INDEX_FILE}" | tr -d ' ')
    log "Wrote ${OUTPUT} (${package_count} Go packages)"
}

main "$@"
