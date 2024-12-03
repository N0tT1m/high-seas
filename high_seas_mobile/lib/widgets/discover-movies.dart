import 'package:flutter/material.dart';
import 'package:high_seas_mobile/widgets/drawer.dart';

class DiscoverMovies extends StatefulWidget {
  const DiscoverMovies({super.key, required this.title});

  final String title;

  @override
  State<DiscoverMovies> createState() => _DiscoverMoviesState();
}

class _DiscoverMoviesState extends State<DiscoverMovies> {
  final List<String> _data = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          backgroundColor: Theme.of(context).colorScheme.inversePrimary,
          title: Text(widget.title),
          leading: Builder(
            builder: (BuildContext context) {
              return IconButton(
                icon: const Icon(Icons.menu),
                onPressed: () { Scaffold.of(context).openDrawer(); },
                tooltip: MaterialLocalizations.of(context).openAppDrawerTooltip,
              );
            },
          ),
        ),
        drawer: const DrawerWidget(),
        body: Center(

          child: SizedBox(
              width: MediaQuery.of(context).size.width,
              height: MediaQuery.of(context).size.height,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    "Discover Movies",
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontFamily: 'JetBrainsMono',
                      fontStyle: FontStyle.normal,
                      fontSize: 28,
                      color: Theme.of(context).colorScheme.onSurface,
                    ),
                  ),
                    SizedBox(
                      width: MediaQuery.of(context).size.width - 120,
                      height: MediaQuery.of(context).size.height - 138,
                      child: ListView.builder(
                        itemCount: _data.length,
                        itemBuilder: (context, index) {
                          return ListTile(
                            title: Text(_data[index]),
                          );
                        },
                      ),
                    ),
                ],
              ),
            ),
        ),
    );
  }
}
