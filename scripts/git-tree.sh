require_clean_work_tree () {
	git rev-parse --verify HEAD >/dev/null || exit 1
	git update-index -q --ignore-submodules --refresh
	err=0
	if ! git diff-files --quiet --ignore-submodules
	then
		action=$1
		case "$action" in
		rebase)
			gettextln "Cannot rebase: You have unstaged changes." >&2
			;;
		"rewrite branches")
			gettextln "Cannot rewrite branches: You have unstaged changes." >&2
			;;
		"pull with rebase")
			gettextln "Cannot pull with rebase: You have unstaged changes." >&2
			;;
		*)
			eval_gettextln "Cannot \$action: You have unstaged changes." >&2
			;;
		esac
		err=1
	fi
	if ! git diff-index --cached --quiet --ignore-submodules HEAD --
	then
		if test $err = 0
		then
			action=$1
			case "$action" in
			rebase)
				gettextln "Cannot rebase: Your index contains uncommitted changes." >&2
				;;
			"pull with rebase")
				gettextln "Cannot pull with rebase: Your index contains uncommitted changes." >&2
				;;
			*)
				eval_gettextln "Cannot \$action: Your index contains uncommitted changes." >&2
				;;
			esac
		else
		    gettextln "Additionally, your index contains uncommitted changes." >&2
		fi
		err=1
	fi
	if test $err = 1
	then
		test -n "$2" && echo "$2" >&2
		exit 1
	fi
}