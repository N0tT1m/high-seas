import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'High Seas',
      theme: ThemeData(
        // This is the theme of your application.
        //
        // TRY THIS: Try running your application with "flutter run". You'll see
        // the application has a purple toolbar. Then, without quitting the app,
        // try changing the seedColor in the colorScheme below to Colors.green
        // and then invoke "hot reload" (save your changes or press the "hot
        // reload" button in a Flutter-supported IDE, or press "r" if you used
        // the command line to start the app).
        //
        // Notice that the counter didn't reset back to zero; the application
        // state is not lost during the reload. To reset the state, use hot
        // restart instead.
        //
        // This works for code too, not just values: Most code changes can be
        // tested with just a hot reload
        colorScheme: ColorScheme(
          surface: Colors.deepPurple.shade900,
          brightness: Brightness.dark,
          primary: const Color.fromRGBO(229, 226, 233, 100),
          secondary: Colors.deepPurpleAccent,
          onPrimary: Colors.black12,
          onSecondary: Colors.cyanAccent,
          error: Colors.red.shade300,
          onError: Colors.teal.shade900,
          onSurface: Color.fromRGBO(253,245,201, 100),
        ),
        useMaterial3: true,
      ),
      home: const MyHomePage(title: 'High Seas Home Page'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // TRY THIS: Try changing the color here to a specific color (to
        // Colors.amber, perhaps?) and trigger a hot reload to see the AppBar
        // change color while the other colors stay the same.
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
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
      drawer: SizedBox(
        width: MediaQuery.of(context).size.width * 0.75, // 75% of screen will be occupied
        child: Drawer(
          backgroundColor: Color.fromRGBO(28, 0, 64, 100),
            // Your drawer content here
            child: ListView(
              padding: EdgeInsets.zero,
              children: [
                DrawerHeader(
                  child: Image.asset(
                    "assets/high-seas-1.jpg",
                    width: MediaQuery.of(context).size.width * 500,
                    height: MediaQuery.of(context).size.height * 500,
                  ),
                  decoration: BoxDecoration(
                    color: Theme.of(context).colorScheme.surface,
                  ),
                ),
                ListTile(
                  tileColor: Theme.of(context).colorScheme.surface,
                  title: Text(
                      "Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 32,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                  ),
                ),
                ListTile(
                  tileColor: Theme.of(context).colorScheme.surface,
                  title: Text(
                    "Discover Movies",
                    style: TextStyle(
                      fontFamily: 'JetBrainsMono',
                      fontStyle: FontStyle.normal,
                      fontSize: 18,
                      color: Theme.of(context).colorScheme.onSurface,
                    ),
                  ),
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => Placeholder(),
                      ),
                    );
                  }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Search Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Now Playing Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Popular Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Top Rated Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Upcoming Movies",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                  tileColor: Theme.of(context).colorScheme.surface,
                  title: Text(
                    "Shows",
                    style: TextStyle(
                      fontFamily: 'JetBrainsMono',
                      fontStyle: FontStyle.normal,
                      fontSize: 32,
                      color: Theme.of(context).colorScheme.onSurface,
                    ),
                  ),
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Discover Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Search Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Airing Today Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Popular Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "Top Rated Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
                ListTile(
                    tileColor: Theme.of(context).colorScheme.surface,
                    title: Text(
                      "On The Air Shows",
                      style: TextStyle(
                        fontFamily: 'JetBrainsMono',
                        fontStyle: FontStyle.normal,
                        fontSize: 18,
                        color: Theme.of(context).colorScheme.onSurface,
                      ),
                    ),
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Placeholder(),
                        ),
                      );
                    }
                ),
              ],
            ),
          ),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Column(
          // Column is also a layout widget. It takes a list of children and
          // arranges them vertically. By default, it sizes itself to fit its
          // children horizontally, and tries to be as tall as its parent.
          //
          // Column has various properties to control how it sizes itself and
          // how it positions its children. Here we use mainAxisAlignment to
          // center the children vertically; the main axis here is the vertical
          // axis because Columns are vertical (the cross axis would be
          // horizontal).
          //
          // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
          // action in the IDE, or press "p" in the console), to see the
          // wireframe for each widget.
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text(
              'You have pushed the button this many times:',
            ),
          ],
        ),
      ),
    );
  }
}
