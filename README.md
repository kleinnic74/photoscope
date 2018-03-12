# Poor man's Photo Management

DIY Photo Management software written in golang, targetted at managing larger collections of photos.

Provides the back end as REST service in Golang and the front-end as an HTML application.

## Goals

* self-contained photo-indexing service capable of handling large collections (>100k images)
* no dependency on external cloud services, all running locally or self-hosted services
* able to run on NAS consumer hardware (e.g. Synology 2xx series)
* native network (REST) API to access all functions

## Roadmap

Current Phase: early prototypes

* basic photo indexing capabilities:
** import photos
** extract time and location meta-data
** index by time, by location
* HTML browser
** browse index
** view photos


Next Phase: more advanced indexing

* temporal event splitting (based on https://www.fxpal.com/publications/temporal-event-clustering-for-digital-photo-collections-2.pdf)
* duplicate detection

Further away:

* face recognition and indexing
* messaging service integration (e.g. Telegram auto-index bot)

## Out of scope

* photo editing
