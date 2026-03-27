# devops-elective-itmo-2026-1lab

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=000000)](https://golang.org/)
[![Linux](https://img.shields.io/badge/Linux-required-FF9900?style=for-the-badge&logo=linux&logoColor=white&labelColor=000000)](https://www.kernel.org/)
[![CI Status](https://img.shields.io/github/actions/workflow/status/apartapatia/devops-elective-itmo-2026-1lab/ci.yml?style=for-the-badge&label=apartapatia-runtime%20CI&color=2ea44f&logo=github&logoColor=white&labelColor=000000)](https://github.com/apartapatia/devops-elective-itmo-2026-1lab/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/badge/Tests-5%20suites-6866cc?style=for-the-badge&labelColor=000000)](https://github.com/apartapatia/devops-elective-itmo-2026-1lab/actions/workflows/ci.yml)
# Отчет по работе


## Список вопросов для самостоятельного исследования и ответов

запуск как parent потом как child?

можно ли считывать namespace прямо из runc?

```
		"namespaces": [
			{
				"type": "pid"
			},
			{
				"type": "network"
			},
			{
				"type": "ipc"
			},
			{
				"type": "uts"
			},
			{
				"type": "mount"
			},
			{
				"type": "cgroup"
			}
		],
```
вот часть runc, в которой выписаны нужные неймспейсы. можем ли просто считать их и не хардкодить их,или будет меняться значение постоянно в спеке или не быть главных в спеке...


defer нужно ли через него что-то чистить после остановки контейнеров


chroot vs pivot_root очень интересно но не понятно)

![](images/pivot_root.svg)

# Иллюстрация к работе кода

![](images/scheme.svg)

# Источники

https://github.com/lizrice/containers-from-scratch/blob/master/main.go

https://www.youtube.com/watch?v=8fi7uSYlOdc

https://github.com/jvns </3
